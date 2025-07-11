<?php

use app\widgets\ScrollWidget;
use app\widgets\SearchResultsSummary;
use src\forms\SearchForm;
use src\helpers\SearchResultHelper;
use src\helpers\TextProcessor;
use src\models\Paragraph;
use src\repositories\ParagraphDataProvider;
use yii\bootstrap5\ActiveForm;
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

/** Search settings form block */
echo Html::beginForm(['/site/search-settings'], 'post', ['name' => 'searchSettingsForm', 'class' => 'd-flex']);
echo Html::hiddenInput('value', 'toggle');
echo Html::endForm();
$inputTemplate = '<div class="input-group mb-2">
          {input}
          <button class="btn btn-primary" type="submit" id="button-search">Поиск</button>
          <button class="btn btn-outline-secondary ' .
  (Yii::$app->session->get('show_search_settings') ? 'active' : "") . '" id="button-search-settings">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-sliders" viewBox="0 0 16 16">
              <path fill-rule="evenodd" d="M11.5 2a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3zM9.05 3a2.5 2.5 0 0 1 4.9 0H16v1h-2.05a2.5 2.5 0 0 1-4.9 0H0V3h9.05zM4.5 7a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3zM2.05 8a2.5 2.5 0 0 1 4.9 0H16v1H6.95a2.5 2.5 0 0 1-4.9 0H0V8h2.05zm9.45 4a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3zm-2.45 1a2.5 2.5 0 0 1 4.9 0H16v1h-2.05a2.5 2.5 0 0 1-4.9 0H0v-1h9.05z"/>
            </svg>
          </button>
          </div>';

?>
<div class="site-index">
  <?php if (Yii::$app->params['cleanDesign']): ?>
    <div class="d-print-none search-block" style="top:0">
    <?php else: ?>
      <div class="search-block">
      <?php endif; ?>
      <div class="container-fluid">

        <?php $form = ActiveForm::begin(
          [
            'method' => 'GET',
            'action' => ['site/index'],
            'options' => ['class' => 'pb-1 mb-2 pt-3', 'autocomplete' => 'off'],
          ]
        ); ?>
        <?= Html::hiddenInput('page', Yii::$app->request->get('page', 1)) ?>
        <div class="d-flex align-items-center">
          <?= $form->field($model, 'query', [
            'inputTemplate' => $inputTemplate,
            'options' => [
              'class' => 'w-100',
              'role' => 'search'
            ]
          ])->textInput(
            [
              'type' => 'search',
              'class' => 'form-control form-control-lg',
              'placeholder' => "Поиск",
              'autocomplete' => 'off',
            ]
          )->label(false); ?>
        </div>
        <div id="search-setting-panel"
          class="search-setting-panel <?= Yii::$app->session->get('show_search_settings') ? 'show-search-settings' : '' ?>">

          <?= $form->field($model, 'matching', ['inline' => true, 'options' => ['class' => 'pb-2']])
            ->radioList($model->getMatching(), ['class' => 'form-check-inline'])
            ->label(false); ?>

          <!-- Чекбокс для включения/выключения однострочного режима -->
          <?php echo $form->field($model, 'singleLineMode', [
            'options' => ['class' => 'pb-3 single-line-mode'],
            'template' => "<div class=\"form-check form-switch\">\n{input}\n{label}\n</div>",
            'labelOptions' => ['class' => 'form-check-label'],
          ])->checkbox([
            'class' => 'form-check-input',
            'id' => 'single-line-mode',
            'uncheck' => null,
            'data-scroll' => 'true', // Добавляем атрибут для обработки скролла
          ], false)->label('Однострочный режим (убрать переносы строк)');
          ?>

        </div>

        <?php ActiveForm::end(); ?>
      </div>
      </div>
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
          $pagination = new Pagination(
            [
              'totalCount' => $results->getTotalCount(),
              'defaultPageSize' => Yii::$app->params['searchResults']['pageSize'],
              'params' => array_merge(Yii::$app->request->get(), [
                'search' => [
                  'singleLineMode' => $model->singleLineMode,
                  'matching' => $model->matching,
                  'query' => $model->query,
                ]
              ]),
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

                <div class="card pt-3 mb-2">
                  <div class="card-body">
                    <?php foreach ($paragraphs as $paragraph): ?>
                      <div class="px-xl-5 px-lg-5 px-md-5 px-sm-3 paragraph" data-entity-id="<?= $paragraph->id; ?>">
                        <div class="paragraph-header">
                          <div class="d-flex justify-content-between">
                            <div>

                            </div>
                            <div class="paragraph-context d-print-none">
                              <?php $total = ceil($paragraph->position / $pagination->pageSize); ?>
                              <?= Html::a(
                                'контекст',
                                [
                                  'site/context',
                                  'id' => $paragraph->id,
                                  'page' => $total,
                                  'f' => $paragraph->position,
                                  '#' => $paragraph->position
                                ],
                                [
                                  'class' => 'btn btn-link btn-context paragraph-context',
                                  'target' => '_blank'
                                ]
                              ); ?>
                            </div>
                          </div>
                        </div>
                        <div>
                          <div class="paragraph-text">
                            <?= SearchResultHelper::highlightFieldContent(
                              $paragraph,
                              'text',
                              'markdown',
                              $model->singleLineMode
                            ); ?>
                          </div>
                        </div>
                        <div class="d-flex justify-content-start book-name">
                          <div><strong><i><?= SearchResultHelper::highlightFieldContent($paragraph, 'genre'); ?>.
                                <?= SearchResultHelper::highlightFieldContent($paragraph, 'author'); ?> —
                                <?= SearchResultHelper::highlightFieldContent($paragraph, 'title'); ?></i></strong></div>
                        </div>
                      </div>
                    <?php endforeach; ?>
                  </div>
                </div>

              <?php endif; ?>
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
      </div>
      <?php if (Yii::$app->params['cleanDesign']): ?>
        <svg xmlns="http://www.w3.org/2000/svg" class="d-none d-print-none">
          <symbol id="check2" viewBox="0 0 16 16">
            <path d="M13.854 3.646a.5.5 0 0 1 0 .708l-7 7a.5.5 0 0 1-.708 0l-3.5-3.5a.5.5 0 1 1 .708-.708L6.5 10.293l6.646-6.647a.5.5 0 0 1 .708 0z"></path>
          </symbol>
          <symbol id="circle-half" viewBox="0 0 16 16">
            <path d="M8 15A7 7 0 1 0 8 1v14zm0 1A8 8 0 1 1 8 0a8 8 0 0 1 0 16z"></path>
          </symbol>
          <symbol id="moon-stars-fill" viewBox="0 0 16 16">
            <path d="M6 .278a.768.768 0 0 1 .08.858 7.208 7.208 0 0 0-.878 3.46c0 4.021 3.278 7.277 7.318 7.277.527 0 1.04-.055 1.533-.16a.787.787 0 0 1 .81.316.733.733 0 0 1-.031.893A8.349 8.349 0 0 1 8.344 16C3.734 16 0 12.286 0 7.71 0 4.266 2.114 1.312 5.124.06A.752.752 0 0 1 6 .278z"></path>
            <path d="M10.794 3.148a.217.217 0 0 1 .412 0l.387 1.162c.173.518.579.924 1.097 1.097l1.162.387a.217.217 0 0 1 0 .412l-1.162.387a1.734 1.734 0 0 0-1.097 1.097l-.387 1.162a.217.217 0 0 1-.412 0l-.387-1.162A1.734 1.734 0 0 0 9.31 6.593l-1.162-.387a.217.217 0 0 1 0-.412l1.162-.387a1.734 1.734 0 0 0 1.097-1.097l.387-1.162zM13.863.099a.145.145 0 0 1 .274 0l.258.774c.115.346.386.617.732.732l.774.258a.145.145 0 0 1 0 .274l-.774.258a1.156 1.156 0 0 0-.732.732l-.258.774a.145.145 0 0 1-.274 0l-.258-.774a1.156 1.156 0 0 0-.732-.732l-.774-.258a.145.145 0 0 1 0-.274l.774-.258c.346-.115.617-.386.732-.732L13.863.1z"></path>
          </symbol>
          <symbol id="sun-fill" viewBox="0 0 16 16">
            <path d="M8 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8zM8 0a.5.5 0 0 1 .5.5v2a.5.5 0 0 1-1 0v-2A.5.5 0 0 1 8 0zm0 13a.5.5 0 0 1 .5.5v2a.5.5 0 0 1-1 0v-2A.5.5 0 0 1 8 13zm8-5a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1 0-1h2a.5.5 0 0 1 .5.5zM3 8a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1 0-1h2A.5.5 0 0 1 3 8zm10.657-5.657a.5.5 0 0 1 0 .707l-1.414 1.415a.5.5 0 1 1-.707-.708l1.414-1.414a.5.5 0 0 1 .707 0zm-9.193 9.193a.5.5 0 0 1 0 .707L3.05 13.657a.5.5 0 0 1-.707-.707l1.414-1.414a.5.5 0 0 1 .707 0zm9.193 2.121a.5.5 0 0 1-.707 0l-1.414-1.414a.5.5 0 0 1 .707-.707l1.414 1.414a.5.5 0 0 1 0 .707zM4.464 4.465a.5.5 0 0 1-.707 0L2.343 3.05a.5.5 0 1 1 .707-.707l1.414 1.414a.5.5 0 0 1 0 .708z"></path>
          </symbol>
        </svg>
        <div class="color-theme-widget dropdown position-fixed bottom-0 end-0 mb-3 me-3 bd-mode-toggle d-print-none">
          <button class="btn btn-secondary py-2 dropdown-toggle d-flex align-items-center" id="bd-theme" type="button" aria-expanded="false" data-bs-toggle="dropdown" aria-label="Toggle theme (light)">
            <svg class="bi my-1 theme-icon-active" width="1em" height="1em">
              <use href="#sun-fill"></use>
            </svg>
            <span class="visually-hidden" id="bd-theme-text">Toggle theme</span>
          </button>
          <ul class="dropdown-menu dropdown-menu-end shadow" aria-labelledby="bd-theme-text">
            <li>
              <button type="button" class="dropdown-item d-flex align-items-center active" data-bs-theme-value="light" aria-pressed="true">
                <svg class="bi me-2 opacity-50 theme-icon" width="1em" height="1em">
                  <use href="#sun-fill"></use>
                </svg>
                Light
                <svg class="bi ms-auto d-none" width="1em" height="1em">
                  <use href="#check2"></use>
                </svg>
              </button>
            </li>
            <li>
              <button type="button" class="dropdown-item d-flex align-items-center" data-bs-theme-value="dark" aria-pressed="false">
                <svg class="bi me-2 opacity-50 theme-icon" width="1em" height="1em">
                  <use href="#moon-stars-fill"></use>
                </svg>
                Dark
                <svg class="bi ms-auto d-none" width="1em" height="1em">
                  <use href="#check2"></use>
                </svg>
              </button>
            </li>
            <li>
              <button type="button" class="dropdown-item d-flex align-items-center" data-bs-theme-value="auto" aria-pressed="false">
                <svg class="bi me-2 opacity-50 theme-icon" width="1em" height="1em">
                  <use href="#circle-half"></use>
                </svg>
                Auto
                <svg class="bi ms-auto d-none" width="1em" height="1em">
                  <use href="#check2"></use>
                </svg>
              </button>
            </li>
          </ul>
        </div>
      <?php endif; ?>

      <?= ScrollWidget::widget(['data_entity_id' => isset($paragraph) ? $paragraph->id : 0]); ?>
    <?php else: ?>
      <!--        <div class="card welcome-card">-->
      <!--          <div class="card-body">-->
      <!--          </div>-->
      <!--        </div>-->
    <?php endif; ?>
    </div>
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
JS;

    $this->registerJs($js);
    ?>