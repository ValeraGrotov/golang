package pfp

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
)

var supportedCharacters = []string{
	"[", "]", "_", "$", " ", "=", ">", "<", "(", ")",
}
var supportedVariablesWithKeywords = map[string]string{
	"$account_id":       "[in]",
	"$answer_option_id": "[in]",
	"$contingent_id":    "[in]",
	"$langcodes":        "[in]",

	"$account_ids":       "[contains]",
	"$answer_option_ids": "[contains]",
	"$contingent_ids":    "[contains]",
}
var supportedKeywords = map[string]string{
	"contains": "ok",
	"in":       "ok",
	"and":      "ok",
	"or":       "ok",
	"=":        "ok",
	">":        "ok",
	"<":        "ok",
}

func removeSpaces(str string) string {
	space, err := regexp.Compile(`\s+`)
	if err != nil {
		log.Println(err)
		return strings.TrimSpace(str)
	}
	str = space.ReplaceAllString(str, " ")
	space, err = regexp.Compile(`\]\s+\[`)
	if err != nil {
		log.Println(err)
		return strings.TrimSpace(str)
	}
	str = space.ReplaceAllString(str, "][")
	return strings.TrimSpace(str)
}

func syntaxValidation(queryString string) error {
	saveQueryString := queryString
	if countOpen, countClose := strings.Count(queryString, "["), strings.Count(queryString, "]"); countOpen > countClose {
		return fmt.Errorf("syntax error: missing pair for %q", "[")
	} else if countOpen < countClose {
		return fmt.Errorf("syntax error: missing pair for %q", "]")
	}

	if countOpen, countClose := strings.Count(queryString, "("), strings.Count(queryString, ")"); countOpen > countClose {
		return fmt.Errorf("syntax error: missing pair for %q", "(")
	} else if countOpen < countClose {
		return fmt.Errorf("syntax error: missing pair for %q", ")")
	}

	if strings.Contains(queryString, "[ ") || strings.Contains(queryString, " ]") {
		return fmt.Errorf("syntax error: do not use spaces inside %q", "[]")
	}
	for _, symbol := range supportedCharacters {
		queryString = strings.Replace(queryString, symbol, "", -1)
	}
	if !govalidator.IsAlphanumeric(queryString) {
		return fmt.Errorf("syntax error: non-latin characters are not supported")
	}

	queryString = saveQueryString

	words := strings.Split(queryString, " ")
	for i := 0; i < len(words); i++ {
		words[i] = strings.TrimSpace(words[i])
		if strings.Contains(words[i], "$") {
			if i+1 == len(words) {
				if supportedVariablesWithKeywords[words[i]] == "" {
					return fmt.Errorf("syntax error: unknown variable %q", words[i])
				}
				continue
			}
			variable := words[i]
			keyword := strings.TrimSpace(words[i+1])

			if supportedVariablesWithKeywords[variable] == "" {
				return fmt.Errorf("syntax error: unknown variable %q", variable)
			}

			if supportedKeywords[keyword] != "ok" {
				return fmt.Errorf("syntax error: unknown keyword %q", keyword)
			}

			if !strings.Contains(supportedVariablesWithKeywords[variable], "["+keyword+"]") {
				return fmt.Errorf("syntax error: can't use keyword %q with variable %q", strings.ToUpper(keyword), variable)
			}
			continue
		}
		if strings.Contains(words[i], "[") || strings.Contains(words[i], "]") {
			if strings.HasPrefix(words[i], "[") && strings.HasSuffix(words[i], "]") {
				varsCount := strings.Count(words[i], "][") + 1
				if varsCount > 0 {
					if strings.Count(words[i], "]") > varsCount || strings.Count(words[i], "[") > varsCount {
						return fmt.Errorf("syntax error: invalid value %q", words[i])
					}
				}
				continue
			} else {
				return fmt.Errorf("syntax error: invalid value %q", words[i])
			}
		}
		if supportedKeywords[words[i]] != "ok" {
			return fmt.Errorf("syntax error: unknown keyword %q", words[i])
		}
	}

	return nil
}

func Parse(queryString string) error {
	queryString = removeSpaces(queryString)
	queryString = strings.ToLower(queryString)
	if err := syntaxValidation(queryString); err != nil {
		return err
	}
	return nil
}
