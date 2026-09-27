package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func ReadJSON(w http.ResponseWriter, r *http.Request, dst any, limit int) error {
	r.Body = http.MaxBytesReader(w, r.Body, int64(limit))
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		var syntaxErr *json.SyntaxError
		var invalidUnmarshalErr *json.InvalidUnmarshalError
		var unMarshalTypeErr *json.UnmarshalTypeError
		var maxByteErr *http.MaxBytesError
		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxErr.Offset)
		case errors.As(err, &invalidUnmarshalErr):
			return err
		case errors.As(err, &unMarshalTypeErr):
			if unMarshalTypeErr.Field != "" {
				return fmt.Errorf("body contains badly-formed JSON type (at character %q)", unMarshalTypeErr.Field)
			}
			return fmt.Errorf("body contains badly-formed JSON type (at character %d)", unMarshalTypeErr.Offset)
		case errors.Is(err, io.EOF):
			return fmt.Errorf("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			field := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", field)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &maxByteErr):
			return fmt.Errorf("body must not be larger than %d bytes", maxByteErr.Limit)
		default:
			return err
		}
	}
	err := decoder.Decode(struct{}{})
	if !errors.Is(err, io.EOF) {
		return fmt.Errorf("body must contains single JSON")
	}
	return nil
}
