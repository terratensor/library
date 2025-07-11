<?php

declare(strict_types=1);

namespace src\helpers;

class SearchHelper
{
    public static array $charactersList = ['!', '"', '$', "'", '(', ')', '-', '/', '<', '@', '\\', '^', '|', '~'];

    /**
     * @param string $queryString
     * @return string
     * Escaping characters in query string
     * As some characters are used as operators in the query string, they should be escaped to avoid query errors
     * or unwanted matching conditions. The following characters should be escaped using backslash (\):
     * https://manual.manticoresearch.com/Searching/Full_text_matching/Escaping
     */
    public static function escapingCharacters(string $queryString): string
    {
        $escapedString = '';
        foreach (str_split($queryString) as $char) {
            foreach (self::$charactersList as $character) {
                if ($char === $character) {
                    $escapedCharacter = '\\' . $character;
                    $queryString = str_replace($character, $escapedCharacter, $queryString);
                    $char = $escapedCharacter;
                }
            }
            $escapedString .= $char;
        }

        // var_dump($queryString);
        return $escapedString;
    }


    public static function containsURL(string $input)
    {
        // Регулярное выражение для поиска URL-адресов в строке
        $pattern = "/(https?:\/\/[^\s]+)/";
        if (preg_match($pattern, $input)) {
            return true;
        } else {
            return false;
        }
    }

    public static function processStringWithURLs(string $input)
    {
        // Регулярное выражение для поиска URL-адресов в строке
        $pattern = "/(https?:\/\/[^\s]+)/";
        preg_match_all($pattern, $input, $matches);

        foreach ($matches[0] as $url) {
            // Экранируем специальные символы в URL
            $escapedURL = self::escapingCharacters($url);
            // Заменяем исходный URL на экранированную версию в строке
            $input = str_replace($url, $escapedURL, $input);
        }

        return $input;
    }

    /**
     * Transforms a string from latin to cyrillic using a predefined mapping
     *
     * @param string $input the string to be transformed
     * @return string the transformed string   
     */
    public static function transformString($input)
    {
        $mapping = [
            'a' => 'ф',
            'b' => 'и',
            'c' => 'с',
            'd' => 'в',
            'e' => 'у',
            'f' => 'а',
            'g' => 'п',
            'h' => 'р',
            'i' => 'ш',
            'j' => 'о',
            'k' => 'л',
            'l' => 'д',
            'm' => 'ь',
            'n' => 'т',
            'o' => 'щ',
            'p' => 'з',
            'q' => 'й',
            'r' => 'к',
            's' => 'ы',
            't' => 'е',
            'u' => 'г',
            'v' => 'м',
            'w' => 'ц',
            'x' => 'ч',
            'y' => 'н',
            'z' => 'я',
            '`' => 'ё',
            '[' => 'х',
            ']' => 'ъ',
            ',' => 'б',
            '.' => 'ю',
            ';' => 'ж',
            '\'' => 'э'
        ];

        $output = '';
        for ($i = 0; $i < strlen($input); $i++) {
            $char = strtolower(substr($input, $i, 1));
            if (isset($mapping[$char])) {
                $output .= $mapping[$char];
            } else {
                $output .= $char;
            }
        }

        return $output;
    }

    /**
     * Escapes unclosed double quotes in a string, so that ManticoreSearch can't confuse them with its own
     * @param string $string
     * @return string
     */
    public static function escapeUnclosedQuotes($string)
    {
        $currently_open = '';
        $position = 0;
        $strLength = strlen($string);

        // Loop through each character in the string
        for ($i = 0; $i < $strLength; $i++) {

            // Skip over escaped double quotes, i.e. \" does not count as an unclosed double quote
            if (substr($string, $i, 2) == "\\\"") {
                $i++;
                continue;
            }

            // $string = self::replaceAsterisk($string, $i);

            // If we encounter a double quote, and we are not currently inside a double quote
            // (i.e. we are not currently counting it as an unclosed double quote), then mark the current
            // position as an unclosed double quote
            if (substr($string, $i, 1) === "\"") {
                if ($currently_open === '') {
                    $currently_open = substr($string, $i, 1);
                    $position = $i;
                } else {
                    $currently_open = '';
                }
            }

            // TODO добавить обработку REGEX operator, чтобы можно было использовать астериск вместе с оператором
            // if (substr($string, $i, 1) === "*" && $currently_open === "") {
            //     $string = self::replaceAsterisk($string, $i);
            // } elseif (substr($string, $i, 1) === "*" && $currently_open === "\"") {
            //     $asteriskPosition = $i;
            // }
        }

        // If we have an unclosed double quote, add an escape character before it, so that
        // ManticoreSearch can't confuse it with its own syntax
        if ($currently_open !== "") {
            $string = substr_replace($string, '\\', $position, -$strLength - $position);
            // TODO добавить обработку REGEX operator, чтобы можно было использовать астериск вместе с оператором
            // $string = self::replaceAsterisk($string, $asteriskPosition);
        }
        // echo $string;
        return $string;
    }

    /**
     * Replaces asterisk (*) in the string if it is not surrounded by alphanumeric characters.
     * На текущий момент не используется 
     * TODO: дописать так, чтобы астериск не экранировался, в том случае если as an any-term modifier within a phrase search.
     * Т.е. внутри фразы - строки, которая обрамлена ковычками. Например: "управление * системами"
     * @param string $string The input string.
     * @param int $i The position of the asterisk in the string.
     * @return string The modified string.
     */
    public static function replaceAsterisk($string, $i)
    {
        // Get the previous and next characters around the asterisk
        $prevChar = substr($string, $i - 1, 1);
        $nextChar = substr($string, $i + 1, 1);

        // Check if the asterisk is not surrounded by alphanumeric characters
        if (!preg_match('/[a-zA-Zа-яА-Я0-9]/u', $prevChar) && !preg_match('/[a-zA-Zа-яА-Я0-9]/u', $nextChar)) {
            // Replace the asterisk with two backslashes to escape it
            $string = str_replace('*', '\\', $string);
        }

        return $string;
    }

    /**
     * Checks if the given query string contains a regex pattern.
     *
     * This function checks if the given query string contains a regex pattern
     * in the form of "REGEX(pattern)". If it does, it returns true, otherwise false.
     *
     * @param string $queryString The query string to check.
     * @return bool True if the query string contains a regex pattern, false otherwise.
     */
    public static function containsRegexPattern(string $queryString): bool
    {
        $regexPattern = '/REGEX\((.+)\)/';
        return preg_match($regexPattern, $queryString) > 0;
    }

    /**
     * Checks if the given query string contains special characters.
     *
     * This function checks if the given query string contains any of the special
     * characters listed in the $charactersList property. If it does, it returns true,
     * otherwise false.
     *
     * @param string $queryString The query string to check.
     * @return bool True if the query string contains special characters, false otherwise.
     */
    public static function containsSpecialChars(string $queryString): bool
    {
        foreach (self::$charactersList as $character) {
            if (strpos($queryString, $character) !== false) {
                return true;
            }
        }
        return false;
    }
}
