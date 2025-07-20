<?php

use app\widgets\ScrollWidget;
use app\widgets\SearchResultsSummary;
use src\forms\SearchForm;
use src\helpers\SearchHelper;
use src\helpers\SearchResultHelper;
use src\helpers\TextProcessor;
use src\models\Paragraph;
use src\repositories\ParagraphDataProvider;
use yii\bootstrap5\ActiveForm;
use yii\bootstrap5\Breadcrumbs;
use yii\bootstrap5\Html;
use yii\bootstrap5\LinkPager;
use yii\data\Pagination;
use yii\helpers\Markdown;

/** @var yii\web\View $this
 * @var ParagraphDataProvider $results
 * @var Pagination $pages
 * @var SearchForm $model
 * @var string $errorQueryMessage
 */

$this->params['aggs'] = $results->responseData ??  [];

$this->title = Yii::$app->name;
$this->params['breadcrumbs'][] = Yii::$app->name;

$this->params['meta_description'] = 'Цитаты из 11 тысяч томов преимущественно русскоязычных авторов, в которых широко раскрыты большинство исторических событий — это документальная, научная, историческая литература, а также воспоминания, мемуары, дневники и письма, издававшиеся в форме собраний сочинений и художественной литературы';

if ($results) {
  $this->registerMetaTag(['name' => 'robots', 'content' => 'noindex, nofollow']);
} else {
  $this->registerLinkTag(['rel' => 'canonical', 'href' => Yii::$app->params['frontendHostInfo']]);
  $this->registerMetaTag(['name' => 'robots', 'content' => 'index, nofollow']);
}

/** Quote form block  */

echo Html::beginForm(['/site/quote'], 'post', ['name' => 'QuoteForm',  'target' => "print_blank"]);
echo Html::hiddenInput('uuid', '', ['id' => 'quote-form-uuid']);
echo Html::endForm();

?>
<div class="site-index">
  <?= $this->render('_search-panel', ['model' => $model]); ?>
  <div class="container-fluid search-results">

    <?php if (!$results): ?>

      <?php if ($errorQueryMessage): ?>
        <div class="card border-danger mb-3">
          <div class="card-body"><?= $errorQueryMessage; ?></div>
        </div>
      <?php endif; ?>

    <?php endif; ?>

    <?php if ($results): ?>
      <?php
      // Property totalCount пусто пока не вызваны данные модели getModels(),
      // сначала получаем массив моделей, потом получаем общее их количество
      /** @var Paragraph[] $paragraphs */
      $paragraphs = $results->getModels();
      $queryParams = Yii::$app->request->queryParams;
      $pagination = new Pagination(
        [
          'totalCount' => $results->getTotalCount(),
          'defaultPageSize' => Yii::$app->params['searchResults']['pageSize'],
        ]
      );
      ?>
      <div class="row">
        <div class="col-md-12">
          <?php if ($pagination->totalCount === 0): ?>
            <h5>По вашему запросу ничего не найдено</h5>
          <?php else: ?>
            <div class="row">
              <div class="col-md-8 d-flex align-items-center">
                <?= SearchResultsSummary::widget(['pagination' => $pagination]); ?>
              </div>
            </div>
            <?php foreach ($paragraphs as $paragraph): ?>
              <div class="card mt-4">
                <div class="card-header d-flex justify-content-between">
                  <?= Breadcrumbs::widget([
                    'homeLink' => false,
                    'links' => array_filter([
                      !empty($paragraph->genre) ? [
                        'label' => $paragraph->genre,
                        'url' => SearchHelper::getFilterUrl('genre', $paragraph->genre),
                        'active' => !empty($model->genre) && $model->genre === $paragraph->genre ? ' active-filter' : '',
                        'data-bs-toggle' => 'tooltip',
                        'data-bs-title' => !empty($model->genre) && $model->genre === $paragraph->genre ? 'Нажмите чтобы снять фильтр' : 'Нажмите чтобы фильтровать по жанру'
                      ] : null,
                      !empty($paragraph->author) ? [
                        'label' => $paragraph->author,
                        'url' => SearchHelper::getFilterUrl('author', $paragraph->author),
                        'active' => !empty($model->author) && $model->author === $paragraph->author ? ' active-filter' : '',
                        'data-bs-toggle' => 'tooltip',
                        'data-bs-title' => !empty($model->author) && $model->author === $paragraph->author ? 'Нажмите чтобы снять фильтр' : 'Нажмите чтобы фильтровать по автору'
                      ] : null,
                      !empty($paragraph->title) ? [
                        'label' => $paragraph->title,
                        'url' => SearchHelper::getFilterUrl('title', $paragraph->title),
                        'active' => !empty($model->title) && $model->title === $paragraph->title ? ' active-filter' : '',
                        'data-bs-toggle' => 'tooltip',
                        'data-bs-title' => !empty($model->title) && $model->title === $paragraph->title ? 'Нажмите чтобы снять фильтр' : 'Нажмите чтобы фильтровать по наименованию'
                      ] : null,
                    ]),
                  ]); ?>
                  <div class="paragraph-context d-print-none">
                    <?php $total = ceil($paragraph->chunk / $pagination->pageSize); ?>
                    <?= Html::a(
                      'контекст',
                      [
                        'site/context',
                        'id' => $paragraph->id,
                        'page' => $total,
                        'f' => $paragraph->chunk,
                        '#' => $paragraph->chunk
                      ],
                      [
                        'class' => 'btn btn-link btn-context paragraph-context',
                        'target' => '_blank'
                      ]
                    ); ?>

                  </div>
                </div>
                <div class="card-body">
                  <div class="py-xl-5 py-3 px-xl-5 px-lg-5 px-md-5 px-sm-3 paragraph" data-entity-id="<?= $paragraph->id; ?>">
                    <!-- <h5><?php SearchResultHelper::highlightFieldContent($paragraph, 'title'); ?></h4> -->
                    <div class=" paragraph-text">
                      <?= SearchResultHelper::highlightFieldContent($paragraph, 'content', 'markdown', $model->singleLineMode); ?>
                    </div>
                  </div>
                </div>
                <div class="card-footer">
                  <div class="d-flex justify-content-between align-items-center">
                    <!-- Левый блок с иконками и статистикой -->
                    <div class="icons-stats d-flex align-items-center gap-3">
                      <div class="icons d-print-none d-flex align-items-center gap-2">
                        <i id="bookmark-<?= $paragraph->id ?>" class="bi bi-bookmark" style="font-size: 1.2rem;  margin-top: -8px"
                          data-bs-toggle="tooltip" data-bs-placement="bottom"
                          data-bs-title="Добавить в закладки"
                          data-href="/bookmark?id=<?= $paragraph->id ?>"
                          data-method="post"></i>
                        <i id="share-<?= $paragraph->id ?>" class="bi bi-share" style="font-size: 1.2rem;  margin-top: -8px"
                          data-bs-toggle="tooltip" data-bs-placement="bottom"
                          data-bs-title="Поделиться"></i>
                      </div>
                      <div class="text-muted small" style="line-height: 1.2; padding-top: 2px">
                        Символов: <?= $paragraph->char_count ?>, слов: <?= $paragraph->word_count ?>
                      </div>
                    </div>

                    <!-- Правый блок с источником -->
                    <div class="source d-flex align-items-center gap-2">
                      <span data-bs-toggle="tooltip" data-bs-placement="left"
                        data-bs-title="<?= Html::encode($paragraph->source) ?>"
                        style="line-height: 1.2">
                        Источник
                      </span>
                      <i class="bi bi-clipboard copy-source"
                        style="cursor: pointer; font-size: 1.2rem; margin-top: -8px"
                        data-bs-toggle="tooltip" data-bs-placement="left"
                        data-bs-title="Копировать источник"
                        data-source="<?= Html::encode($paragraph->source) ?>"></i>
                    </div>
                  </div>
                </div>
              </div>
            <?php endforeach; ?>
        </div>
      </div>

    <?php endif; ?>

    <!-- Пагинация -->
    <div class="container container-pagination d-p">
      <div class="detachable">
        <?= LinkPager::widget([
          'pagination' => $pagination,
          'firstPageLabel' => true,
          'lastPageLabel' => false,
          'maxButtonCount' => 5,
          'options' => ['class' => 'd-flex justify-content-center'],
          'listOptions' => ['class' => 'pagination mb-0']
        ]); ?>
      </div>
    </div>

  </div>
</div>

<?= $this->render('_theme-toggler'); ?>
<?= ScrollWidget::widget(['data_entity_id' => isset($paragraph) ? $paragraph->id : 0]); ?>

<?php endif; ?>

<?php $js = <<<JS
let menu = $(".search-block");
var menuOffsetTop = menu.offset().top;
var menuHeight = menu.outerHeight();
var menuParent = menu.parent();
var menuParentPaddingTop = parseFloat(menuParent.css("padding-top"));
 
checkWidth();
 
function checkWidth() {
    if (menu.length !== 0) {
      $(window).scroll(onScroll);
    }
}
 
function onScroll() {
  if ($(window).scrollTop() > menuOffsetTop) {
    menu.addClass("shadow");
    menuParent.css({ "padding-top": menuParentPaddingTop });
  } else {
    menu.removeClass("shadow");
    menuParent.css({ "padding-top": menuParentPaddingTop });
  }
}

const btn = document.getElementById('button-search-settings');
btn.addEventListener('click', toggleSearchSettings, false)

function toggleSearchSettings(event) {
  event.preventDefault();
  btn.classList.toggle('active')
  document.getElementById('search-setting-panel').classList.toggle('show-search-settings')
  
  const formData = new FormData(document.forms.searchSettingsForm);
  let xhr = new XMLHttpRequest();
  xhr.open("POST", "/site/search-settings");
  xhr.send(formData);
}
// Обработчик ссылок контекста
const contextButtons = document.querySelectorAll('button.btn-context')
contextButtons.forEach(function (element) {
  element.addEventListener('click', btnContextHandler, false)
})

function btnContextHandler(event) {
  const quoteForm = document.forms["QuoteForm"]
  const uuid = document.getElementById("quote-form-uuid")
  uuid.value = event.target.dataset.uuid
  quoteForm.submit();
}


$('input[type=radio]').on('change', function() {
    $(this).closest("form").submit();
});

JS;

$this->registerJs($js);
?>

<?php
$js = <<<JS
// Функция для определения видимого параграфа с учетом sticky-панели
function getVisibleParagraphId() {
    const paragraphs = document.querySelectorAll('.paragraph');
    const searchBlock = document.querySelector('.search-block');
    const searchBlockHeight = searchBlock ? searchBlock.offsetHeight : 0;
    
    let visibleParagraphId = null;
    let maxVisibleArea = 0;
    
    paragraphs.forEach(paragraph => {
        const rect = paragraph.getBoundingClientRect();
        // Вычисляем видимую высоту с учетом sticky-панели
        const visibleHeight = Math.min(rect.bottom, window.innerHeight) - 
                             Math.max(rect.top, searchBlockHeight);
        
        if (visibleHeight > 0 && visibleHeight > maxVisibleArea) {
            maxVisibleArea = visibleHeight;
            visibleParagraphId = paragraph.dataset.entityId;
        }
    });
    
    return visibleParagraphId || (paragraphs.length > 0 ? paragraphs[0].dataset.entityId : null);
}

// Функция для скролла к параграфу
function scrollToParagraph() {
    const urlParams = new URLSearchParams(window.location.search);
    const paragraphId = urlParams.get('scrollTo');
    
    if (paragraphId) {
        const element = document.querySelector('.paragraph[data-entity-id="' + paragraphId + '"]');
        if (element) {
            setTimeout(() => {
                // Учитываем высоту sticky-панели при скролле
                const searchBlock = document.querySelector('.search-block');
                const offset = searchBlock ? searchBlock.offsetHeight + 20 : 20;
                const elementPosition = element.getBoundingClientRect().top + window.pageYOffset;
                
                window.scrollTo({
                    top: elementPosition - offset,
                    behavior: 'smooth'
                });
                
                // Подсвечиваем параграф
                element.style.transition = 'background-color 0.5s';
                element.style.backgroundColor = '#f8f9fa';
                
                setTimeout(() => {
                    element.style.backgroundColor = '';
                }, 2000);
            }, 100);
            
            // Удаляем параметр из URL
            urlParams.delete('scrollTo');
            const newUrl = window.location.pathname + '?' + urlParams.toString();
            window.history.replaceState({}, '', newUrl);
        }
    }
}

// Обработчик чекбокса
document.getElementById('single-line-mode').addEventListener('change', function() {
    const visibleParagraphId = getVisibleParagraphId();
    const urlParams = new URLSearchParams(window.location.search);
    
    urlParams.set('search[singleLineMode]', this.checked ? '1' : '0');
    
    if (visibleParagraphId) {
        urlParams.set('scrollTo', visibleParagraphId);
    }
    
    if (urlParams.has('page')) {
        urlParams.set('page', urlParams.get('page'));
    }
    
    window.location.search = urlParams.toString();
});

// Инициализация скролла
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', scrollToParagraph);
} else {
    scrollToParagraph();
}

$('form').on('submit', function() {
    // Сохраняем текущие значения фильтров перед отправкой
    const searchParams = new URLSearchParams(window.location.search);
    const genre = searchParams.get('search[genre]');
    const author = searchParams.get('search[author]');
    const title = searchParams.get('search[title]');
    
    if (genre) {
        $(this).append('<input type="hidden" name="search[genre]" value="' + genre + '">');
    }
    if (author) {
        $(this).append('<input type="hidden" name="search[author]" value="' + author + '">');
    }
    if (title) {
        $(this).append('<input type="hidden" name="search[title]" value="' + title + '">');
    }
});

// Инициализация tooltips для активных фильтров
$(document).ready(function() {
    $('[data-bs-toggle="tooltip"]').tooltip();
});

// Обработчик кликов по кнопкам удаления фильтров
$(document).on('click', '.filter-badge .close', function(e) {
    e.preventDefault();
    window.location.href = $(this).attr('href');
});

   // Обработчик копирования источника
  $('.copy-source').on('click', function() {
    const sourceText = $(this).data('source');
    const tooltip = bootstrap.Tooltip.getInstance(this);
    
    // Создаем временный textarea для копирования
    const textarea = document.createElement('textarea');
    textarea.value = sourceText;
    textarea.style.position = 'fixed';  // Чтобы не было прокрутки страницы
    document.body.appendChild(textarea);
    textarea.select();
    
    try {
      // Пробуем использовать современный API
      if (navigator.clipboard) {
        navigator.clipboard.writeText(sourceText).then(function() {
          showCopiedTooltip(tooltip, this);
        }.bind(this));
      } else {
        // Старый метод для браузеров без Clipboard API
        const success = document.execCommand('copy');
        if (success) {
          showCopiedTooltip(tooltip, this);
        } else {
          throw new Error('Copy command failed');
        }
      }
    } catch (err) {
      console.error('Ошибка копирования:', err);
      // Показываем сообщение об ошибке
      $(this).attr('data-bs-title', 'Ошибка копирования');
      tooltip.show();
      setTimeout(() => {
        $(this).attr('data-bs-title', 'Копировать источник');
        tooltip.hide();
      }, 2000);
    } finally {
      // Удаляем временный элемент
      document.body.removeChild(textarea);
    }
  });
  
  function showCopiedTooltip(tooltip, element) {
    $(element).attr('data-bs-title', 'Скопировано!');
    tooltip.show();
    setTimeout(() => {
      $(element).attr('data-bs-title', 'Копировать источник');
      tooltip.hide();
    }, 2000);
  }

JS;

$this->registerJs($js);
?>