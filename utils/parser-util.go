package utils

import (
	"errors"
	"regexp"
	"unicode"
	"unicode/utf8"
)

// We use content of the "/etc/passwd" file to parse login shell by user name.
// For each user /etc/passwd file stores information in the separated line. Information split with help ":".
// Row information:
// - User name
// - Encrypted password
// - User ID number (UID)
// - User's group ID number (GID)
// - Full name of the user (GECOS)
// - User home directory
// - Login shel
// So, each line starts with username and ends by login shell path.
// Read more: https://www.ibm.com/support/knowledgecenter/en/ssw_aix_72/com.ibm.aix.security/passwords_etc_passwd_file.htm
func ParseEtcPassWd(etcPassWdContent string, userId string) (shell string, err error) {
	rgExp, err := regexp.Compile(".*:.*:" + userId + ":.*:.*:.*:(?P<ShellPath>.*)")
	if err != nil {
		return "", err
	}

	result := rgExp.FindStringSubmatch(etcPassWdContent)
	// first group it's all expression, second on it's "?P<ShellPath>"
	if len(result) != 2 {
		return "", errors.New("unable to find default shell")
	}

	return result[1], nil
}

//Parse user id from not digital characters(new line character, control sequences)
func ParseUID(content []byte) (uid string, err error) {
	var userIdString string
	for offset := 1; offset < len(content); {
		cRune, size := utf8.DecodeRune(content[offset:])
		offset += size
		if unicode.IsDigit(cRune) {
			userIdString += string(cRune)
		}
	}
	return userIdString, nil
}
