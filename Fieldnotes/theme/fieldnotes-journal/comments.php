<?php if (post_password_required()) return; $fieldnotes_has_comments = have_comments(); ?>
<section class="comments-area <?php echo $fieldnotes_has_comments ? 'has-comments' : 'no-comments'; ?>" id="comments">
    <?php if ($fieldnotes_has_comments) : ?><h2 class="section-heading"><?php printf(esc_html(_n('%s 条评论', '%s 条评论', get_comments_number(), 'fieldnotes-journal')), number_format_i18n(get_comments_number())); ?></h2><ol class="comment-list"><?php wp_list_comments(array('style' => 'ol', 'short_ping' => true, 'avatar_size' => 38)); ?></ol><?php the_comments_navigation(); endif; ?>
    <?php if (!comments_open() && get_comments_number()) : ?><p><?php esc_html_e('评论已关闭。', 'fieldnotes-journal'); ?></p><?php endif; ?>
    <?php
    $fieldnotes_form_args = array('title_reply' => __('落笔留声', 'fieldnotes-journal'));
    if (is_singular('post') && comments_open()) {
        ob_start();
        get_template_part('template-parts/article-author', null, array('author' => fieldnotes_article_author(get_post_field('post_author', get_the_ID()))));
        $fieldnotes_form_args['title_reply_before'] = ob_get_clean() . '<h3 id="reply-title" class="comment-reply-title">';
    }
    comment_form($fieldnotes_form_args);
    ?>
</section>
