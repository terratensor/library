<?php

declare(strict_types=1);

namespace src\repositories;


use Manticoresearch\Search;
use src\models\Paragraph;
use yii\data\BaseDataProvider;

class ParagraphDataProvider extends BaseDataProvider
{
    /**
     * @var string|callable Имя столбца с ключом или callback-функция, возвращающие его
     */
    public $key;
    /**
     * @var Search
     */
    public Search $query;

    protected function prepareModels(): array
    {
        $models = [];
        $pagination = $this->getPagination();
        $sort = $this->getSort();

        foreach ($sort->getOrders() as $attribute => $value) {
            $direction = $value === SORT_ASC ? 'asc' : 'desc';
            $this->query->sort($attribute, $direction);
        }

        if ($pagination === false) {
            // в случае отсутствия разбивки на страницы - прочитать все строки
            foreach ($this->query->get() as $hit) {
                $model = new Paragraph($hit->getData());
                $model->setId((int)$hit->getId());
                $models[] = $model;
            }
        } else {
            // в случае, если разбивка на страницы есть - прочитать только одну страницу
            $pagination->totalCount = $this->getTotalCount();

            $this->query->limit($pagination->pageSize);
            $this->query->offset($pagination->getOffset());

            $this->prepareMaxMatches();

            $limit = $pagination->getLimit();

            try {
                $data = $this->query->get();
            } catch (\Exception $e) {
                return [];
            }

            // Если количество записей меньше чем лимит,
            // то переписываем лимит, чтобы избежать ошибки Undefined array key при вызове $data->current()
            if ($data->count() < $limit) {
                $limit = $data->count();
            }

            for ($count = 0; $count < $limit; ++$count) {
                $model = new Paragraph( $data->current()->getData());
                $model->setId((int)$data->current()->getId());
                try {
                    $model->highlight = $data->current()->getHighlight();
                } catch (\Exception $e) {
                    $model->highlight = [];
                }
                $models[] = $model;
                $data->next();
            }
        }

        return $models;
    }

    protected function prepareKeys($models): array
    {
        if ($this->key !== null) {
            $keys = [];

            foreach ($models as $model) {
                if (is_string($this->key)) {
                    $keys[] = $model[$this->key];
                } else {
                    $keys[] = call_user_func($this->key, $model);
                }
            }

            return $keys;
        } else {
            return array_keys($models);
        }
    }

    protected function prepareTotalCount()
    {
        try {
            return $this->query->get()->getTotal();
        } catch (\Exception $e) {
            return 0;
        }
    }

    /**
     * https://manual.manticoresearch.com/Searching/Options#max_matches
     * Если ограничение max_matches установлено, то устанавливаем счетчик кол-во совпадений равный параметру max_matches
     * Иначе устанавливаем значение запроса max_matches равное кол-во существующих результатов, т.е. выдача без ограничения
     * By default, Manticore Search uses a result set window of 1000 best-ranked documents that can be returned in the result set.
     * If the result set is paginated beyond this value, the query will end in error.
     *
     * This limitation can be adjusted with the query option max_matches.
     *
     * Increasing the max_matches to very high values should only be done if it's necessary for the navigation to reach such
     * points. A high max_matches value requires more memory and can increase the query response time. One way to work with deep
     * result sets is to set max_matches as the sum of the offset and limit.
     *
     * Lowering max_matches below 1000 has the benefit of reducing the memory used by the query. It can also reduce the query
     * time, but in most cases, it might not be a noticeable gain.
     */
    private function prepareMaxMatches(): void
    {
        $max_matches = (int)\Yii::$app->params['manticore']['max_matches'];
        if ($this->pagination->getOffset() >= $max_matches) {
            $this->query->maxMatches($this->pagination->getOffset() + $this->pagination->getLimit());
        }
    }
}
