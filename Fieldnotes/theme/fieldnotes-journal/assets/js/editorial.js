/* Author portraits only. Keep WordPress and comment avatar providers intact. */
(function () {
    'use strict';
    function fallback(event) {
        var image = event.target;
        if (image.matches && image.matches('.reading-author-portrait img')) image.remove();
    }
    document.addEventListener('error', fallback, true);
    document.querySelectorAll('.reading-author-portrait img').forEach(function (image) {
        if (image.complete && image.naturalWidth === 0) image.remove();
    });
}());
