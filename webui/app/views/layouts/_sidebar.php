<?php

use yii\helpers\Html;
use yii\helpers\Url;

$aggs = $this->params['aggs'] ?? [];
?>
<?php if (empty($aggs)) return; ?>
<div class="sidebar-wrapper">
    <div class="sidebar-toggle" onclick="toggleSidebar()">
        <i class="bi bi-chevron-double-left"></i>
    </div>

    <div class="sidebar">
        <div class="sidebar-header">
            <h5>Фильтры</h5>
        </div>

        <div class="sidebar-content">
            <div class="total-count mb-3">
                <small>Всего документов:</small>
                <h6><?= number_format($aggs['hits']['total'] ?? 0, 0, '', ' ') ?></h6>
            </div>

            <div class="accordion" id="filtersAccordion">
                <!-- Жанры -->
                <div class="accordion-item">
                    <h2 class="accordion-header">
                        <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#genreCollapse">
                            <i class="bi bi-bookmark me-2"></i> Жанры
                        </button>
                    </h2>
                    <div id="genreCollapse" class="accordion-collapse collapse" data-bs-parent="#filtersAccordion">
                        <div class="accordion-body p-0">
                            <div class="facet-search mb-2 px-2">
                                <input type="text" class="form-control form-control-sm" placeholder="Поиск жанров...">
                            </div>
                            <ul class="facet-list">
                                <?php foreach ($aggs['aggregations']['genre_group']['buckets'] as $genre): ?>
                                    <?php if (!empty($genre['key'])): ?>
                                        <li>
                                            <a href="<?= Url::to(['site/search', 'search' => ['genre' => $genre['key']]]) ?>">
                                                <?= Html::encode($genre['key']) ?>
                                                <span class="badge bg-secondary float-end"><?= number_format($genre['doc_count'], 0, '', ' ') ?></span>
                                            </a>
                                        </li>
                                    <?php endif; ?>
                                <?php endforeach; ?>
                            </ul>
                        </div>
                    </div>
                </div>

                <!-- Авторы -->
                <div class="accordion-item">
                    <h2 class="accordion-header">
                        <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#authorCollapse">
                            <i class="bi bi-person me-2"></i> Авторы
                        </button>
                    </h2>
                    <div id="authorCollapse" class="accordion-collapse collapse" data-bs-parent="#filtersAccordion">
                        <div class="accordion-body p-0">
                            <div class="facet-search mb-2 px-2">
                                <input type="text" class="form-control form-control-sm" placeholder="Поиск авторов...">
                            </div>
                            <ul class="facet-list">
                                <?php foreach ($aggs['aggregations']['author_group']['buckets'] as $author): ?>
                                    <?php if (!empty($author['key'])): ?>
                                        <li>
                                            <a href="<?= Url::to(['site/search', 'search' => ['author' => $author['key']]]) ?>">
                                                <?= Html::encode($author['key']) ?>
                                                <span class="badge bg-secondary float-end"><?= number_format($author['doc_count'], 0, '', ' ') ?></span>
                                            </a>
                                        </li>
                                    <?php endif; ?>
                                <?php endforeach; ?>
                            </ul>
                        </div>
                    </div>
                </div>

                <!-- Названия -->
                <div class="accordion-item">
                    <h2 class="accordion-header">
                        <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#titleCollapse">
                            <i class="bi bi-card-text me-2"></i> Названия
                        </button>
                    </h2>
                    <div id="titleCollapse" class="accordion-collapse collapse" data-bs-parent="#filtersAccordion">
                        <div class="accordion-body p-0">
                            <div class="facet-search mb-2 px-2">
                                <input type="text" class="form-control form-control-sm" placeholder="Поиск названий...">
                            </div>
                            <ul class="facet-list">
                                <?php foreach ($aggs['aggregations']['title_group']['buckets'] as $title): ?>
                                    <?php if (!empty($title['key'])): ?>
                                        <li>
                                            <a href="<?= Url::to(['site/search', 'search' => ['title' => $title['key']]]) ?>">
                                                <?= Html::encode(mb_substr($title['key'], 0, 30) . (mb_strlen($title['key']) > 30 ? '...' : '')) ?>
                                                <span class="badge bg-secondary float-end"><?= number_format($title['doc_count'], 0, '', ' ') ?></span>
                                            </a>
                                        </li>
                                    <?php endif; ?>
                                <?php endforeach; ?>
                            </ul>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>

<script>
    function toggleSidebar() {
        const sidebar = document.querySelector('.sidebar-wrapper');
        sidebar.classList.toggle('collapsed');

        // Сохраняем состояние в localStorage
        const isCollapsed = sidebar.classList.contains('collapsed');
        localStorage.setItem('sidebarCollapsed', isCollapsed);
    }

    // Проверяем состояние при загрузке
    document.addEventListener('DOMContentLoaded', function() {
        const isCollapsed = localStorage.getItem('sidebarCollapsed') === 'true';
        if (isCollapsed) {
            document.querySelector('.sidebar-wrapper').classList.add('collapsed');
        }

        // Поиск внутри фасетов
        document.querySelectorAll('.facet-search input').forEach(input => {
            input.addEventListener('keyup', function() {
                const searchText = this.value.toLowerCase();
                const list = this.closest('.accordion-body').querySelector('.facet-list');

                Array.from(list.querySelectorAll('li')).forEach(li => {
                    const text = li.textContent.toLowerCase();
                    li.style.display = text.includes(searchText) ? '' : 'none';
                });
            });
        });
    });
</script>