<?php if (!defined('ABSPATH')) exit;
if (fieldnotes_show_site_footer()) :
?>
<footer class="site-footer home-colophon">
    <div class="colophon-quiet-inner">
        <p class="colophon-theme"><?php esc_html_e('主题', 'fieldnotes-journal'); ?><span>Fieldnotes Journal</span></p>
        <p class="colophon-copyright">&copy; <?php echo esc_html(wp_date('Y')); ?> <?php echo esc_html(get_bloginfo('name')); ?></p>
    </div>
</footer>
<?php endif; wp_footer(); ?>
</body>
</html>
