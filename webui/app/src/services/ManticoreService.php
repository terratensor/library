<?php

declare(strict_types=1);

namespace src\services;

use src\forms\SearchForm;
use src\repositories\ParagraphDataProvider;
use src\repositories\ParagraphRepository;
use Yii;

class ManticoreService
{
    private ParagraphRepository $paragraphRepository;

    public function __construct(ParagraphRepository $questionRepository)
    {
        $this->paragraphRepository = $questionRepository;
    }

    /**
     * @param SearchForm $form
     * @return ParagraphDataProvider
     * @throws EmptySearchRequestExceptions
     */
    public function search(SearchForm $form): ParagraphDataProvider
    {
        $queryString = $form->query;
        $comments = $this->paragraphRepository->findByQueryStringNew($queryString, $form);

        $responseData = $comments->get()->getResponse()->getResponse();

        return new ParagraphDataProvider(
            [
                'query' => $comments,
                'pagination' => [
                    'pageSize' => Yii::$app->params['searchResults']['pageSize'],
                ],
                'sort' => [
                    //                 'defaultOrder' => [
                    //     'id' => SORT_ASC,
                    //     'chunk' => SORT_ASC,
                    // ],
                    'attributes' => [
                        'id',
                        'chunk',
                    ]
                ],
                'responseData' => $responseData
            ]
        );
    }

    public function aggs(SearchForm $form)
    {
        $resp = $this->paragraphRepository->findAggsAll($form);
        return $resp->getResponse();
    }

    public function findByBook(int $id): ParagraphDataProvider
    {
        $paragraphs = $this->paragraphRepository->findParagraphsByBookId($id);

        return new ParagraphDataProvider(
            [
                'query' => $paragraphs,
                'pagination' => [
                    'pageSize' => Yii::$app->params['searchResults']['pageSize'],
                ],
                'sort' => [
                    'defaultOrder' => [
                        'id' => SORT_ASC,
                        'chunk' => SORT_ASC,
                    ],
                    'attributes' => [
                        'id',
                        'chunk'
                    ]
                ],
            ]
        );
    }

    public function findBook($id): \Manticoresearch\ResultSet
    {
        return $this->paragraphRepository->findBookById((int)$id);
    }
}
