(function () {
  'use strict';
  function el(tag, className, text) {
    var node = document.createElement(tag);
    if (className) node.className = className;
    if (text !== undefined) node.textContent = String(text);
    return node;
  }
  function icon(name) {
    var paths = {
      arrow: 'M6 18 18 6M6 6h12v12',
      landscape: 'M3 18h18M4 16l5-6 4 4 3-3 4 5M15 5h.01',
      search: 'm16 16 5 5M17 10a7 7 0 1 1-14 0 7 7 0 0 1 14 0'
    };
    var svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('viewBox', '0 0 24 24');
    svg.setAttribute('fill', 'none');
    svg.setAttribute('stroke', 'currentColor');
    svg.setAttribute('stroke-width', '1.5');
    svg.setAttribute('stroke-linecap', 'round');
    svg.setAttribute('stroke-linejoin', 'round');
    svg.setAttribute('aria-hidden', 'true');
    svg.setAttribute('focusable', 'false');
    var path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
    path.setAttribute('d', paths[name] || paths.landscape);
    svg.appendChild(path);
    return svg;
  }
  function safeURL(value) {
    if (!value || typeof value !== 'string') return '';
    try {
      var parsed = new URL(value, window.location.href);
      return parsed.protocol === 'https:' || parsed.protocol === 'http:' ? parsed.href : '';
    } catch (error) { return ''; }
  }
  function cover(point, className) {
    var frame = el('span', className);
    frame.appendChild(icon('landscape'));
    var url = safeURL(point.preview);
    if (url) {
      var image = el('img');
      image.src = url; image.alt = ''; image.loading = 'lazy'; image.decoding = 'async';
      image.addEventListener('error', function () { image.remove(); });
      frame.appendChild(image);
    }
    return frame;
  }
  function popupFor(point, labels) {
    var root = el('div', 'fntm-popup');
    var preview = safeURL(point.preview);
    if (preview) {
      var image = el('img', 'fntm-popup-image');
      image.src = preview; image.alt = ''; image.decoding = 'async';
      image.addEventListener('error', function () { image.remove(); });
      root.appendChild(image);
    }
    var copy = el('div', 'fntm-popup-copy');
    copy.appendChild(el('span', 'fntm-popup-kicker', [point.country, point.year].filter(Boolean).join(' · ')));
    copy.appendChild(el('h3', '', point.title));
    if (point.description) copy.appendChild(el('p', 'fntm-popup-description', point.description));
    if (point.stories.length) {
      var list = el('ul', 'fntm-popup-stories');
      point.stories.forEach(function (story) {
        var item = el('li'), link = el('a');
        link.href = story.url;
        link.appendChild(el('span', '', story.title)); link.appendChild(icon('arrow'));
        item.appendChild(link); list.appendChild(item);
      });
      copy.appendChild(list);
    } else copy.appendChild(el('p', 'fntm-popup-empty', labels.noPosts || '笔记待续。'));
    root.appendChild(copy);
    return root;
  }
  function init(wrapper) {
    var node = wrapper.querySelector('.fntm-map'), source = wrapper.querySelector('.fntm-map-data');
    if (!node || !source || wrapper.dataset.mapReady === '1') return;
    var data;
    try { data = JSON.parse(source.textContent); } catch (error) { return; }
    wrapper.dataset.mapReady = '1';
    var config = window.fntmConfig || {}, labels = config.labels || {}, settings = data.settings || {};
    var seen = Object.create(null);
    var points = (Array.isArray(data.places) ? data.places : []).filter(function (point) {
      if (!point || !/^\d+$/.test(String(point.id)) || seen[point.id] || !Number.isFinite(point.lat) || !Number.isFinite(point.lng) || Math.abs(point.lat) > 90 || Math.abs(point.lng) > 180) return false;
      seen[point.id] = true; return true;
    }).map(function (point) {
      return Object.assign({}, point, {
        title: String(point.title || ''), country: String(point.country || ''), year: String(point.year || ''),
        description: String(point.description || ''),
        stories: (Array.isArray(point.stories) ? point.stories : []).filter(function (story) { return story && safeURL(story.url); }).map(function (story) { return { title: String(story.title || ''), url: safeURL(story.url) }; })
      });
    }).sort(function (a, b) { return b.year.localeCompare(a.year, undefined, { numeric: true }) || a.title.localeCompare(b.title); });
    var center = settings.center || config.center || [23, 12], zoom = Number(settings.zoom || config.zoom || 2);
    var countrySelect = wrapper.querySelector('.fntm-filter-country'), yearSelect = wrapper.querySelector('.fntm-filter-year');
    var search = wrapper.querySelector('.fntm-search'), countLabel = wrapper.querySelector('.fntm-map-count');
    var list = wrapper.querySelector('.fntm-place-list'), mapStatus = wrapper.querySelector('.fntm-map-status');
    var statusText = wrapper.querySelector('.fntm-map-status-text'), retry = wrapper.querySelector('.fntm-tile-retry');
    var browse = wrapper.querySelector('.fntm-browse-places'), reset = wrapper.querySelector('.fntm-map-reset');
    var map = null, markerLayer = null, tiles = null, hasLand = false, hasTiles = false, tileDeadline;
    var markers = Object.create(null), activeId = '', selectionEpoch = 0, visible = points;
    var reducedMotion = Boolean(window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches);
    var mapEmpty = el('div', 'fntm-map-empty'); mapEmpty.hidden = true; node.parentNode.appendChild(mapEmpty);

    function state(name, message) {
      wrapper.dataset.mapState = name;
      if (mapStatus) mapStatus.hidden = !message;
      if (statusText) statusText.textContent = message || '';
      if (retry) retry.hidden = !tiles;
      if (browse) browse.hidden = Boolean(map);
    }
    function startTileDeadline() {
      window.clearTimeout(tileDeadline);
      tileDeadline = window.setTimeout(function () {
        if (!hasTiles) state(hasLand ? 'outline' : 'unavailable', hasLand ? '简图模式 · 地点与游记仍可浏览' : '底图暂时未能展开，可以先浏览地点笔记。');
      }, 8000);
    }
    function setActive(id, scroll) {
      activeId = String(id || '');
      wrapper.querySelectorAll('.fntm-place-card').forEach(function (card) {
        var active = card.dataset.placeId === activeId;
        card.classList.toggle('is-active', active);
        card.querySelector('.fntm-place-select').setAttribute('aria-pressed', String(active));
        if (active && scroll && list && list.clientHeight) {
          var top = card.offsetTop, bottom = top + card.offsetHeight;
          if (top < list.scrollTop || bottom > list.scrollTop + list.clientHeight) {
            var target = Math.max(0, top - 8);
            if (list.scrollTo) list.scrollTo({ top: target, behavior: reducedMotion ? 'auto' : 'smooth' });
            else list.scrollTop = target;
          }
        }
      });
      Object.keys(markers).forEach(function (key) {
        var element = markers[key].getElement();
        if (element) element.classList.toggle('is-active', key === activeId);
      });
    }
    function setView(view) {
      wrapper.dataset.view = view === 'list' ? 'list' : 'map';
      wrapper.querySelectorAll('[data-map-view]').forEach(function (button) { button.setAttribute('aria-pressed', String(button.getAttribute('data-map-view') === wrapper.dataset.view)); });
      if (map && wrapper.dataset.view === 'map') map.invalidateSize({ pan: false });
    }
    function selectPoint(point) {
      setActive(point.id, false);
      var marker = markers[String(point.id)];
      if (!map || !marker) return;
      var epoch = ++selectionEpoch;
      var wasList = wrapper.dataset.view === 'list';
      setView('map');
      map.setView([point.lat, point.lng], Math.max(5, map.getZoom()), { animate: false });
      function reveal() {
        if (epoch !== selectionEpoch || markers[String(point.id)] !== marker) return;
        setActive(point.id, false); marker.openPopup();
        var element = marker.getElement();
        if (wasList && element) element.focus({ preventScroll: true });
      }
      if (markerLayer.zoomToShowLayer) markerLayer.zoomToShowLayer(marker, reveal);
      else reveal();
    }
    function makeCard(point) {
      var card = el('article', 'fntm-place-card'); card.dataset.placeId = String(point.id);
      var choose = el('button', 'fntm-place-select'); choose.type = 'button';
      choose.setAttribute('aria-label', '在地图中查看' + point.title + '，' + [point.country, point.year].filter(Boolean).join('，'));
      choose.setAttribute('aria-pressed', 'false');
      choose.appendChild(cover(point, 'fntm-place-thumb'));
      var copy = el('span', 'fntm-place-copy');
      copy.appendChild(el('span', 'fntm-place-country', point.country || 'TRAVEL NOTE'));
      copy.appendChild(el('strong', 'fntm-place-title', point.title));
      copy.appendChild(el('span', 'fntm-place-meta', [point.year, point.stories.length ? point.stories.length + ' 篇笔记' : '笔记待续'].filter(Boolean).join(' · ')));
      choose.appendChild(copy);
      var arrow = icon('arrow'); arrow.setAttribute('class', 'fntm-place-arrow'); choose.appendChild(arrow);
      choose.addEventListener('click', function () { selectPoint(point); }); card.appendChild(choose);
      if (point.stories.length) {
        var link = el('a', 'fntm-place-story', point.stories[0].title); link.href = point.stories[0].url; card.appendChild(link);
      } else card.appendChild(el('span', 'fntm-place-story fntm-place-unlinked', labels.noPosts || '笔记待续。'));
      return card;
    }
    function clearFilters() {
      if (search) search.value = ''; if (countrySelect) countrySelect.value = ''; if (yearSelect) yearSelect.value = '';
      render(true);
    }
    function emptyState() {
      var root = el('div', 'fntm-place-empty'); root.appendChild(icon('search'));
      root.appendChild(el('p', '', points.length ? '这一处记忆，暂时还没找到。' : '旅程尚未落笔。'));
      if (points.length) {
        var button = el('button', 'fntm-filter-reset', '清除筛选'); button.type = 'button';
        button.addEventListener('click', clearFilters); root.appendChild(button);
      }
      return root;
    }
    function fitPoints(force) {
      if (!map || (!force && settings.fitBounds === false)) return;
      if (visible.length > 1) map.fitBounds(visible.map(function (point) { return [point.lat, point.lng]; }), { padding: [42, 42], maxZoom: 6, animate: false });
      else if (visible.length === 1) map.setView([visible[0].lat, visible[0].lng], Math.max(zoom, 7), { animate: false });
      else map.setView(center, zoom, { animate: false });
    }
    function render(shouldFit) {
      ++selectionEpoch; activeId = '';
      if (map) map.closePopup(); if (markerLayer) markerLayer.clearLayers(); markers = Object.create(null);
      if (list) list.replaceChildren();
      var country = countrySelect ? countrySelect.value : '', year = yearSelect ? yearSelect.value : '';
      var query = search ? search.value.trim().toLocaleLowerCase() : '';
      visible = points.filter(function (point) {
        var text = [point.title, point.country, point.year, point.description].concat(point.stories.map(function (story) { return story.title; })).join(' ').toLocaleLowerCase();
        return (!country || point.country === country) && (!year || point.year === year) && (!query || text.indexOf(query) !== -1);
      });
      visible.forEach(function (point) {
        if (list) list.appendChild(makeCard(point));
        if (!map || !markerLayer) return;
        var pin = el('span', point.stories.length ? 'fntm-pin' : 'fntm-pin fntm-pin-gray'); pin.setAttribute('aria-hidden', 'true');
        var marker = window.L.marker([point.lat, point.lng], { title: point.title, alt: point.title, icon: window.L.divIcon({ className: 'fntm-div-icon', html: pin.outerHTML, iconSize: [32, 32], iconAnchor: [16, 16], popupAnchor: [0, -10] }) });
        marker.bindPopup(popupFor(point, labels), { maxWidth: 280, minWidth: 180, maxHeight: Math.max(180, Math.min(380, (node.clientHeight || 500) - 80)), autoPanPadding: [18, 18] });
        marker.on('popupopen', function () { setActive(point.id, true); });
        marker.on('click', function () { ++selectionEpoch; setActive(point.id, true); });
        markers[String(point.id)] = marker; markerLayer.addLayer(marker);
      });
      if (countLabel) countLabel.textContent = visible.length + ' / ' + points.length + ' 处地点';
      mapEmpty.hidden = visible.length > 0; mapEmpty.replaceChildren();
      if (!visible.length) { if (list) list.appendChild(emptyState()); mapEmpty.appendChild(emptyState()); }
      if (shouldFit) fitPoints(false);
    }
    function addOptions(select, values, reverse) {
      if (!select) return;
      values.sort(function (a, b) { return a.localeCompare(b, undefined, { numeric: true }); }); if (reverse) values.reverse();
      values.forEach(function (value) { var option = el('option', '', value); option.value = value; select.appendChild(option); });
    }
    addOptions(countrySelect, Array.from(new Set(points.map(function (point) { return point.country; }).filter(Boolean))), false);
    addOptions(yearSelect, Array.from(new Set(points.map(function (point) { return point.year; }).filter(Boolean))), true);

    try {
      if (window.L) {
        map = window.L.map(node, { scrollWheelZoom: false, worldCopyJump: true, zoomControl: false, minZoom: 1, maxZoom: 19, preferCanvas: false, zoomAnimation: !reducedMotion, fadeAnimation: !reducedMotion }).setView(center, zoom);
        map.attributionControl.setPrefix('<a href="https://leafletjs.com/">Leaflet</a>');
        window.L.control.zoom({ position: 'topright' }).addTo(map);
        if (window.fntmWorldLandPath && window.L.svgOverlay) {
          var pane = map.createPane('fntm-land'); pane.style.zIndex = '190'; pane.style.pointerEvents = 'none';
          var land = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
          land.setAttribute('viewBox', '0 0 1000 1000'); land.setAttribute('aria-hidden', 'true');
          var outline = document.createElementNS('http://www.w3.org/2000/svg', 'path');
          outline.setAttribute('d', window.fntmWorldLandPath); outline.setAttribute('fill', '#fafafa');
          outline.setAttribute('stroke', '#d7d8dc'); outline.setAttribute('stroke-width', '0.7');
          outline.setAttribute('vector-effect', 'non-scaling-stroke'); land.appendChild(outline);
          window.L.svgOverlay(land, [[-85.05112878, -180], [85.05112878, 180]], { pane: 'fntm-land', interactive: false }).addTo(map);
          hasLand = true;
          map.attributionControl.addAttribution('Land: <a href="https://www.naturalearthdata.com/">Natural Earth</a>');
        }
        var url, options = { maxZoom: 19, className: 'fntm-tile-layer' };
        if (config.token) {
          var style = String(config.style || 'mapbox/light-v11').replace(/[^A-Za-z0-9_/-]/g, ''); if (style.indexOf('/') < 0) style = 'mapbox/light-v11';
          url = 'https://api.mapbox.com/styles/v1/' + style + '/tiles/256/{z}/{x}/{y}?access_token=' + encodeURIComponent(config.token);
          options.attribution = '© <a href="https://www.mapbox.com/about/maps/">Mapbox</a> © <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>';
        } else {
          url = 'https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png'; options.subdomains = 'abcd'; options.detectRetina = true;
          options.attribution = '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap contributors</a> © <a href="https://carto.com/attributions">CARTO</a>';
        }
        tiles = window.L.tileLayer(url, options);
        tiles.on('tileload', function () { hasTiles = true; window.clearTimeout(tileDeadline); state('ready', ''); });
        tiles.on('tileerror', function () { if (!hasTiles) state(hasLand ? 'outline' : 'unavailable', hasLand ? '简图模式 · 地点与游记仍可浏览' : '底图暂时未能展开，可以先浏览地点笔记。'); });
        tiles.addTo(map);
        markerLayer = settings.cluster !== false && window.L.markerClusterGroup ? window.L.markerClusterGroup({ maxClusterRadius: Number(config.clusterRadius || 55), showCoverageOnHover: false, spiderfyOnMaxZoom: true, animate: !reducedMotion, animateAddingMarkers: false }) : window.L.layerGroup();
        markerLayer.addTo(map); state(hasLand ? 'outline' : 'loading', ''); startTileDeadline();
      } else state('unavailable', '地图暂时未能展开，可以先浏览地点笔记。');
    } catch (error) {
      if (map) { map.remove(); map = null; markerLayer = null; tiles = null; }
      state('unavailable', '地图暂时未能展开，可以先浏览地点笔记。');
    }
    if (countrySelect) countrySelect.addEventListener('change', function () { render(true); });
    if (yearSelect) yearSelect.addEventListener('change', function () { render(true); });
    if (search) {
      search.addEventListener('input', function () { render(false); });
      search.addEventListener('keydown', function (event) { if (event.key === 'Escape' && search.value) { search.value = ''; render(true); } });
    }
    wrapper.querySelectorAll('[data-map-view]').forEach(function (button) { button.addEventListener('click', function () { setView(button.getAttribute('data-map-view')); }); });
    if (reset) reset.addEventListener('click', function () { ++selectionEpoch; setActive('', false); if (map) map.closePopup(); fitPoints(true); });
    if (browse) browse.addEventListener('click', function () { setView('list'); });
    if (retry) retry.addEventListener('click', function () { if (!tiles) return; hasTiles = false; state(hasLand ? 'outline' : 'loading', ''); tiles.redraw(); startTileDeadline(); });
    function resize() { if (map && wrapper.dataset.view !== 'list') map.invalidateSize({ pan: false }); }
    if (map && window.ResizeObserver) { var observer = new window.ResizeObserver(resize); observer.observe(node); }
    else if (map) window.addEventListener('resize', resize);
    render(true); window.setTimeout(resize, 100);
    window.addEventListener('pagehide', function () { window.clearTimeout(tileDeadline); });
  }
  function start() { document.querySelectorAll('.fntm-map-wrap').forEach(init); }
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', start); else start();
}());
