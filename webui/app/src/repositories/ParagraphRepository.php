<?php

declare(strict_types=1);

namespace src\repositories;

use Yii;
use src\forms\SearchForm;
use src\models\Paragraph;
use Manticoresearch\Table;
use Manticoresearch\Client;
use Manticoresearch\Search;
use Manticoresearch\Query\In;
use src\helpers\SearchHelper;
use Manticoresearch\Query\Equals;
use Manticoresearch\Query\BoolQuery;
use Manticoresearch\Query\MatchQuery;
use Manticoresearch\Query\MatchPhrase;
use Manticoresearch\Query\QueryString;

class ParagraphRepository
{
    private Client $client;
    public Table $table;
    private Search $search;

    private string $indexName = 'library2025';
    public int $pageSize = 20;

    public function __construct(Client $client, $pageSize)
    {
        $this->setIndexName(\Yii::$app->params['indexes']['common']);
        $this->client = $client;
        $this->search = new Search($this->client);
        $this->search->setTable($this->indexName);
        $this->pageSize = $pageSize;
    }

    /**
     * @param string $queryString
     * @param string|null $indexName
     * @param SearchForm|null $form
     * @return Search
     * "query_string" accepts an input string as a full-text query in MATCH() syntax
     */
    public function findByQueryStringNew(
        string $queryString,
        ?string $indexName = null,
        ?SearchForm $form = null
    ): Search {
        // $this->search->reset();
        if ($indexName) {
            // $this->setIndex($this->search->setTable($indexName));
        }

        $queryString = SearchHelper::processStringWithURLs($queryString);
        $queryString = SearchHelper::escapeUnclosedQuotes($queryString);

        // Запрос переделан под фильтр
        $query = new BoolQuery();

        if ($form->query) {
            $query->must(new QueryString($queryString));
        }

        // Выполняем поиск если установлен фильтр или установлен строка поиска
        if ($form->query) {
            $search = $this->search->search($query);
        } else {
            throw new \DomainException('Задан пустой поисковый запрос');
        }

        // Если нет совпадений no_match_size возвращает пустое поле для подсветки
        $search->highlight(
            ['genre', 'author', 'title', 'text'],
            [
                'limit' => 0,
                'no_match_size' => 0,
                'pre_tags' => '<mark>',
                'post_tags' => '</mark>'
            ],
        );

        if ($form->genre !== '') {
            $search->filter('genre', 'in', $form->genre);
        }

        if ($form->author !== '') {
            $search->filter('author', 'in', $form->author);
        }

        if ($form->title !== '') {
            $search->filter('title_attr', 'in', $form->title);
        }

        return $search;
    }

    public function matchAll()
    {
        $search = $this->client->search(['body' => ['table' => 'library2025', 'query' => ['match_all' => '']]], true);
        return $search;
    }

    /**
     * @param string $queryString
     * @param string|null $indexName
     * @param SearchForm|null $form
     * @return Search
     * "match" is a simple query that matches the specified keywords in the specified fields.
     */
    public function findByQueryStringMatch(
        string $queryString,
        ?string $indexName = null,
        ?SearchForm $form = null
    ): Search {
        $this->search->reset();
        if ($indexName) {
            // $this->setIndex($this->client->index($indexName));
        }

        // Запрос переделан под фильтр
        $query = new BoolQuery();

        if ($form->query) {
            $query->must(new MatchQuery($queryString, '*'));
        }

        // Выполняем поиск если установлен фильтр или установлен строка поиска
        if ($form->query) {
            $search = $this->search->search($query);
        } else {
            throw new \DomainException('Задан пустой поисковый запрос');
        }

        $search->highlight(
            ['genre', 'author', 'title', 'text'],
            [
                'limit' => 0,
                'no_match_size' => 0,
                'pre_tags' => '<mark>',
                'post_tags' => '</mark>'
            ]
        );

        return $search;
    }

    /**
     * @param string $queryString
     * @param string|null $indexName
     * @return Search
     * "match_phrase" is a query that matches the entire phrase. It is similar to a phrase operator in SQL.
     */
    public function findByMatchPhrase(
        string $queryString,
        ?string $indexName = null,
        ?SearchForm $form = null
    ): Search {
        $this->search->reset();
        if ($indexName) {
            // $this->setIndex($this->client->index($indexName));
        }

        // Запрос переделан под фильтр
        $query = new BoolQuery();

        if ($form->query) {
            $query->must(new MatchPhrase($queryString, '*'));
        }


        // Выполняем поиск если установлен фильтр или установлен строка поиска
        if ($form->query) {
            $search = $this->search->search($query);
        } else {
            throw new \DomainException('Задан пустой поисковый запрос');
        }

        $search->highlight(
            ['genre', 'author', 'title', 'text'],
            [
                'limit' => 0,
                'no_match_size' => 0,
                'pre_tags' => '<mark>',
                'post_tags' => '</mark>'
            ]
        );

        return $search;
    }

    /**
     * @param SearchForm $form
     * @param string|null $indexName
     * @return Search
     * "match" is a query that matches the entire phrase. It is similar to a phrase operator in SQL.
     * The search is carried out by genre, author, title
     */
    public function findByContext(SearchForm $form, ?string $indexName = null): Search
    {
        $this->search->reset();
        if ($indexName) {
            // $this->setIndex($this->client->index($indexName));
        }

        // Запрос переделан под фильтр
        $query = new BoolQuery();

        $query->must(new Equals('source_uuid', $form->source_uuid));

        $search = $this->search->search($query);
        $search->facet('source_uuid');

        $search->highlight(
            ['genre', 'author', 'title', 'text'],
            [
                'limit' => 0,
                'no_match_size' => 0,
                'pre_tags' => '<mark>',
                'post_tags' => '</mark>'
            ]
        );

        return $search;
    }

    /**
     * @param $queryString String Число или строка чисел через запятую
     * @param string|null $indexName
     * @return Search
     * Поиск по data_id, вопрос или комментарий, число или массив data_id
     */
    public function findByParagraphId(
        string $queryString,
        ?string $indexName = null,
        ?SearchForm $form = null
    ): Search {
        $this->search->reset();
        if ($indexName) {
            // $this->setIndex($this->client->index($indexName));
        }

        $result = explode(',', $queryString);

        foreach ($result as $key => $item) {
            $item = (int)$item;
            if ($item == 0) {
                unset($result[$key]);
                continue;
            }
            $result[$key] = $item;
        }
        // Запрос переделан под фильтр
        $query = new BoolQuery();

        if (!empty($result)) {
            $query->must(new In('id', array_values($result)));
        } else {
            throw new \DomainException('Неправильный запрос, при поиске по номеру(ам) надо указать номер параграфа, или перечислить номера через запятую');
        }

        // Выполняем поиск если установлен фильтр или установлен строка поиска
        if ($form->query) {
            $search = $this->search->search($query);
        } else {
            throw new \DomainException('Задан пустой поисковый запрос');
        }

        $search->highlight(
            ['text'],
            [
                'limit' => 0,
                'no_match_size' => 0,
                'pre_tags' => '<mark>',
                'post_tags' => '</mark>'
            ]
        );
        $search->sort('id', 'asc');

        return $search;
    }

    /**
     * @param Index $index
     */
    // public function setIndex(Index $index): void
    // {
    //     $this->index = $index;
    // }

    public function findBookById(int $id): \Manticoresearch\ResultSet
    {
        $this->search->reset();

        $search = $this->search->setTable($this->indexName);

        $search->search('');
        $search->filter('book_id', $id);
        $search->limit(1);

        return $search->get();
    }

    public function findParagraphsByBookId(int $id): Search
    {
        $this->search->reset();

        $search = $this->search->setTable($this->indexName);


        $search->filter('book_id', 'in', $id);

        // Запрос переделан под фильтр
        $query = new BoolQuery();


        $query->must(new In('id', $id));
        $this->search->search($query);

        $search->highlight(
            ['text'],
            [
                'limit' => 0,
                'no_match_size' => 0,
                'pre_tags' => '<mark>',
                'post_tags' => '</mark>'
            ]
        );
        return $search;
    }

    /**
     * @param string $indexName
     */
    public function setIndexName(string $indexName): void
    {
        $this->indexName = $indexName;
    }

    /**
     * Возвращает paragraph по uuid
     * @deprecated
     * @param string $uuid
     * @return Paragraph
     */
    public function findByParagraphUuid(string $uuid): Paragraph
    {

        $this->search->reset();

        $query = new BoolQuery();

        $query->must(new Equals('uuid', $uuid));

        $search = $this->search->search($query);

        $current = $search->get()->current();
        $paragraph = new Paragraph($current->getData());
        $paragraph->setId((int)$current->getId());
        return $paragraph;
    }

    /**
     * Возвращает paragraph по id
     * @param string $uuid
     * @return Paragraph
     */
    public function getByParagraphID(string $id): Paragraph
    {
        $table =  new \Manticoresearch\Table($this->client, 'library2025');
        /** @var \Manticoresearch\ResultHit **/
        $hit = $table->getDocumentById($id);

        if (!$hit) {
            throw new \DomainException('Параграф с не найден');
        }

        $par = new Paragraph($hit->getData());
        $par->setId((int)$hit->getId());

        return $par;
    }
}
