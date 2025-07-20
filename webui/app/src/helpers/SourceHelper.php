<?php 

declare(strict_types=1);

namespace src\helpers;

use src\models\Paragraph;

class SourceHelper
{
    public static function fullName(Paragraph $paragraph): string
    {
        $name = "";
        if ($paragraph->genre) {
            $name .= $paragraph->genre.". ";
        }       
        if ($paragraph->author && $paragraph->title) {
            $name .= $paragraph->author ." — " . $paragraph->title;
        } 
        if ($paragraph->author && !$paragraph->title) {
            $name .= $paragraph->author;
        }
        if ($paragraph->title && !$paragraph->author) {
            $name .= $paragraph->title;
        }
        return $name;
    }
}