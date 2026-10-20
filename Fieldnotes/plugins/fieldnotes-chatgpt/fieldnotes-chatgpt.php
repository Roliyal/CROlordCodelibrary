<?php
/**
 * Plugin Name: Fieldnotes ChatGPT Connector
 * Description: A private Echoes workspace with OpenAI chat, attachments, media and remote MCP tools.
 * Version: 1.5.0
 * Requires at least: 6.2
 * Requires PHP: 7.4
 * Author: OpenAI
 * License: GPL-2.0-or-later
 * Text Domain: fieldnotes-chatgpt
 */
if (!defined('ABSPATH')) { exit; }

define('FIELDNOTES_CHATGPT_VERSION', '1.5.0');
define('FIELDNOTES_CHATGPT_FILE', __FILE__);
define('FIELDNOTES_CHATGPT_DIR', plugin_dir_path(__FILE__));
define('FIELDNOTES_CHATGPT_URL', plugin_dir_url(__FILE__));

require_once FIELDNOTES_CHATGPT_DIR . 'includes/class-fieldnotes-chatgpt-connection.php';
require_once FIELDNOTES_CHATGPT_DIR . 'includes/class-fieldnotes-chatgpt-workspace.php';
require_once FIELDNOTES_CHATGPT_DIR . 'includes/class-fieldnotes-chatgpt.php';
Fieldnotes_ChatGPT::instance();

register_activation_hook(__FILE__, array('Fieldnotes_ChatGPT', 'activate'));
