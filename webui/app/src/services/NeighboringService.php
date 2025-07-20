<?php

declare(strict_types=1);

namespace src\services;

use src\forms\QuoteForm;
use src\forms\SearchForm;
use src\models\QuoteResultPdo;
use src\repositories\ParagraphRepository;

class NeighboringService
{
    private ParagraphRepository $paragraphRepository;

    public function __construct(ParagraphRepository $paragraphRepository)
    {
        $this->paragraphRepository = $paragraphRepository;
    }

    public function handle(QuoteForm $quoteForm): QuoteResultPdo
    {
        $paragraph = $this->paragraphRepository->findByParagraphUuid($quoteForm->uuid);
        $form = new SearchForm();
        $form->matching = 'in';
        $form->query = implode(',', $this->getList((int)$paragraph->getId(), 3));

        return new QuoteResultPdo($paragraph->book_name, $form);
    }

    public function getList(int $paragraphID, int $num): array
    {
        $forward = $paragraphID;
        $backward = $paragraphID;
        $forwardList = [];
        $backwardList = [];
        for ($n = 1; $n <= $num; $n++) {
            if ($backward > 1) {
                $backwardList[] = --$backward;
            }
            $forwardList[] = ++$forward;
        }

        $result = array_merge($backwardList, [$paragraphID], $forwardList);
        asort($result);
        return $result;
    }
}
