(function () {
  'use strict';
  document.documentElement.classList.add('js-ready');
  document.documentElement.classList.add('fn-focus-ready');

  // Typing into a clicked field should not turn on a heavy keyboard outline.
  document.addEventListener('keydown', function (event) {
    if (event.key === 'Tab' && !event.altKey && !event.ctrlKey && !event.metaKey) document.documentElement.classList.add('fn-keyboard-focus');
  }, true);
  function usePointerFocus() { document.documentElement.classList.remove('fn-keyboard-focus'); }
  document.addEventListener('pointerdown', usePointerFocus, true);
  document.addEventListener('mousedown', usePointerFocus, true);
  document.addEventListener('touchstart', usePointerFocus, { capture: true, passive: true });

  var header = document.querySelector('.site-header');
  var reduceMotion = window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  function updateHeader() {
    if (header) header.classList.toggle('is-scrolled', window.scrollY > (document.body.classList.contains('single-post') ? 12 : 8));
  }
  updateHeader();
  window.addEventListener('scroll', updateHeader, { passive: true });

  // Native details keeps the single search usable even when JavaScript is off.
  if (header) {
    var navigation = header.querySelector('.header-navigation');
    var menuToggle = header.querySelector('.menu-toggle');
    var siteSearch = header.querySelector('.site-search');
    var compactHeader = window.matchMedia ? window.matchMedia('(max-width: 1080px)') : null;
    function setMenu(open, restoreFocus) {
      if (!navigation || !menuToggle) return;
      var expanded = !!(open && compactHeader && compactHeader.matches);
      navigation.hidden = !!(compactHeader && compactHeader.matches && !expanded);
      menuToggle.setAttribute('aria-expanded', String(expanded));
      header.classList.toggle('menu-is-open', expanded);
      if (restoreFocus) menuToggle.focus();
    }
    function closeSearch(restoreFocus) {
      if (!siteSearch || !siteSearch.open) return;
      siteSearch.open = false;
      if (restoreFocus) siteSearch.querySelector('summary').focus();
    }
    if (navigation && menuToggle && compactHeader) {
      menuToggle.hidden = false;
      setMenu(false);
      menuToggle.addEventListener('click', function () {
        var open = menuToggle.getAttribute('aria-expanded') !== 'true';
        closeSearch(false);
        setMenu(open);
        if (open) {
          var firstLink = navigation.querySelector('a');
          if (firstLink) firstLink.focus();
        }
      });
      navigation.addEventListener('click', function (event) {
        if (event.target.closest('a')) setMenu(false);
      });
      function resetMenu() { setMenu(false); }
      if (compactHeader.addEventListener) compactHeader.addEventListener('change', resetMenu);
      else if (compactHeader.addListener) compactHeader.addListener(resetMenu);
    }
    if (siteSearch) {
      var searchClose = siteSearch.querySelector('.search-close');
      if (searchClose) {
        searchClose.hidden = false;
        searchClose.addEventListener('click', function () { closeSearch(true); });
      }
      siteSearch.addEventListener('toggle', function () {
        if (!siteSearch.open) return;
        setMenu(false);
        window.requestAnimationFrame(function () {
          if (siteSearch.open) siteSearch.querySelector('.search-field').focus({ preventScroll: true });
        });
      });
      siteSearch.addEventListener('focusout', function (event) {
        if (event.relatedTarget && !siteSearch.contains(event.relatedTarget)) closeSearch(false);
      });
    }
    document.addEventListener('click', function (event) {
      if (siteSearch && !siteSearch.contains(event.target)) closeSearch(false);
      if (!header.contains(event.target)) setMenu(false);
    });
    document.addEventListener('keydown', function (event) {
      if (event.key !== 'Escape') return;
      if (siteSearch && siteSearch.open) {
        event.preventDefault();
        closeSearch(true);
      } else if (menuToggle && menuToggle.getAttribute('aria-expanded') === 'true') {
        event.preventDefault();
        setMenu(false, true);
      }
    });
  }

  // Include announcements, wrapped branding and the admin bar in the first view.
  var home = document.querySelector('.page-home');
  var hero = home && home.querySelector('.home-hero');
  if (hero) {
    var homeLayoutFrame;
    function measureHome() {
      if (homeLayoutFrame) window.cancelAnimationFrame(homeLayoutFrame);
      homeLayoutFrame = window.requestAnimationFrame(function () {
        var top = Math.max(0, hero.getBoundingClientRect().top + window.scrollY);
        home.style.setProperty('--home-top', top + 'px');
      });
    }
    measureHome();
    window.addEventListener('resize', measureHome, { passive: true });
    window.addEventListener('load', measureHome, { once: true });
    if (document.fonts && document.fonts.ready) document.fonts.ready.then(measureHome);
    if ('ResizeObserver' in window) {
      var homeLayoutObserver = new ResizeObserver(measureHome);
      [header, document.querySelector('.notice-bar'), document.getElementById('wpadminbar')].forEach(function (node) {
        if (node) homeLayoutObserver.observe(node);
      });
    }
  }

  document.addEventListener('click', function (event) {
    var link = event.target.closest('a[href^="#"]');
    if (!link) return;
    var id = link.getAttribute('href').slice(1);
    if (!id) return;
    var target = document.getElementById(decodeURIComponent(id));
    if (!target) return;
    event.preventDefault();
    target.scrollIntoView({ behavior: reduceMotion ? 'auto' : 'smooth', block: 'start' });
  });

  var revealNodes = Array.prototype.slice.call(document.querySelectorAll('.home-hero-copy,.home-hero-visual,.home-section-head,.post-card,.sidebar-block,.category-card,.fave-card,.entry-featured'));
  if (!reduceMotion && 'IntersectionObserver' in window) {
    var revealObserver = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (!entry.isIntersecting) return;
        entry.target.classList.add('is-revealed');
        revealObserver.unobserve(entry.target);
      });
    }, { threshold: 0.12, rootMargin: '0px 0px -36px 0px' });
    revealNodes.forEach(function (node, index) {
      node.classList.add('reveal-up');
      node.style.transitionDelay = Math.min(index % 4, 3) * 45 + 'ms';
      revealObserver.observe(node);
    });
  }

  var button = document.getElementById('fieldnotes-load-more');
  var list = document.getElementById('fieldnotes-post-list');
  if (button && list && window.fieldnotesConfig) {
    button.addEventListener('click', function () {
      var nextPage = Number(list.dataset.page || 1) + 1;
      var maxPages = Number(list.dataset.maxPages || 1);
      if (nextPage > maxPages || button.disabled) return;
      button.disabled = true;
      button.setAttribute('aria-busy', 'true');
      button.textContent = window.fieldnotesConfig.loadingLabel;
      var data = new FormData();
      data.append('action', 'fieldnotes_load_more');
      data.append('page', String(nextPage));
      data.append('nonce', window.fieldnotesConfig.nonce);
      fetch(window.fieldnotesConfig.ajaxUrl, { method: 'POST', credentials: 'same-origin', body: data })
        .then(function (response) { return response.json(); })
        .then(function (result) {
          if (!result || !result.success) throw new Error('load failed');
          var parsed = new DOMParser().parseFromString(result.data.html, 'text/html');
          Array.prototype.forEach.call(parsed.querySelectorAll('.post-card'), function (card) {
            list.appendChild(card);
            if (!reduceMotion && 'IntersectionObserver' in window) {
              card.classList.add('reveal-up');
              window.requestAnimationFrame(function () { card.classList.add('is-revealed'); });
            }
          });
          list.dataset.page = String(result.data.page);
          if (!result.data.hasMore) button.remove();
          else {
            button.disabled = false;
            button.removeAttribute('aria-busy');
            button.textContent = window.fieldnotesConfig.moreLabel;
          }
        })
        .catch(function () {
          button.disabled = false;
          button.removeAttribute('aria-busy');
          button.textContent = window.fieldnotesConfig.moreLabel;
        });
    });
  }
}());
