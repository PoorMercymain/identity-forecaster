package json_duplicate_checker

import (
	"encoding/json"

	jsonErrors "identity-forecaster/json-errors"
)

func CheckDuplicatesInJSON(d *json.Decoder, path []string) error {
	t, err := d.Token()
	if err != nil {
		return err
	}

	delim, ok := t.(json.Delim)

	if !ok {
		return nil
	}

	if delim == '{' {
		keys := make(map[string]bool)
		for d.More() {
			t, err := d.Token()
			if err != nil {
				return err
			}
			key := t.(string)

			if keys[key] {
				return jsonErrors.ErrDuplicateFieldInJSON
			}

			keys[key] = true

			if err := CheckDuplicatesInJSON(d, append(path, key)); err != nil {
				return err
			}
		}

		if _, err := d.Token(); err != nil {
			return err
		}
	}

	return nil
}
