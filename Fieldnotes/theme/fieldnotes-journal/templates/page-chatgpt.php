<?php
/**
 * Template Name: ChatGPT 对话
 * Description: Password-gated API chat page supplied by Fieldnotes ChatGPT Connector.
 */
if (!defined('ABSPATH')) { exit; }
get_header();
?>
<main id="main-content" class="page-shell conversation-page">
    <div class="chatgpt-page">
        <?php while (have_posts()) : the_post(); ?>
            <header class="chatgpt-page-heading">
                <h1 class="page-title"><?php the_title(); ?></h1>
                <p class="page-description"><?php esc_html_e('让思绪有来处，也有回应。', 'fieldnotes-journal'); ?></p>
            </header>
            <div class="chatgpt-page-content">
                <?php the_content(); ?>
            </div>
        <?php endwhile; ?>
    </div>
</main>
<?php get_footer(); ?>
