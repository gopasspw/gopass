/*
 * Copyright (c) 2013-2014 Conformal Systems <info@conformal.com>
 *
 * This file originated from: http://opensource.conformal.com/
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

#pragma once

#include <stdint.h>
#include <stdlib.h>
#include <string.h>

static GtkAboutDialog *
toGtkAboutDialog(void *p)
{
	return (GTK_ABOUT_DIALOG(p));
}

static GtkAppChooser *
toGtkAppChooser(void *p)
{
	return (GTK_APP_CHOOSER(p));
}

static GtkAppChooserButton *
toGtkAppChooserButton(void *p)
{
	return (GTK_APP_CHOOSER_BUTTON(p));
}

static GtkAppChooserDialog *
toGtkAppChooserDialog(void *p)
{
	return (GTK_APP_CHOOSER_DIALOG(p));
}

static GtkAppChooserWidget *
toGtkAppChooserWidget(void *p)
{
	return (GTK_APP_CHOOSER_WIDGET(p));
}

static GtkApplication *
toGtkApplication(void *p)
{
	return (GTK_APPLICATION(p));
}

static GtkApplicationWindow *
toGtkApplicationWindow(void *p)
{
	return (GTK_APPLICATION_WINDOW(p));
}

static GtkAssistant *
toGtkAssistant(void *p)
{
	return (GTK_ASSISTANT(p));
}

static GtkCalendar *
toGtkCalendar(void *p)
{
	return (GTK_CALENDAR(p));
}

static GtkColorChooserDialog *
toGtkColorChooserDialog(void *p)
{
	return (GTK_COLOR_CHOOSER_DIALOG(p));
}

static GtkDrawingArea *
toGtkDrawingArea(void *p)
{
	return (GTK_DRAWING_AREA(p));
}

static GtkCellRendererSpinner *
toGtkCellRendererSpinner(void *p)
{
	return (GTK_CELL_RENDERER_SPINNER(p));
}

static GtkEventBox *
toGtkEventBox(void *p)
{
	return (GTK_EVENT_BOX(p));
}

static GtkGrid *
toGtkGrid(void *p)
{
	return (GTK_GRID(p));
}

static GtkWidget *
toGtkWidget(void *p)
{
	return (GTK_WIDGET(p));
}

static GtkContainer *
toGtkContainer(void *p)
{
	return (GTK_CONTAINER(p));
}

static GtkOverlay *
toGtkOverlay(void *p)
{
	return (GTK_OVERLAY(p));
}

static GtkPageSetup *
toGtkPageSetup(void *p)
{
	return (GTK_PAGE_SETUP(p));
}

static GtkPaned *
toGtkPaned(void *p)
{
	return (GTK_PANED(p));
}

static GtkPrintContext *
toGtkPrintContext(void *p)
{
	return (GTK_PRINT_CONTEXT(p));
}

static GtkPrintOperation *
toGtkPrintOperation(void *p)
{
	return (GTK_PRINT_OPERATION(p));
}

static GtkPrintOperationPreview *
toGtkPrintOperationPreview(void *p)
{
	return (GTK_PRINT_OPERATION_PREVIEW(p));
}

static GtkPrintSettings *
toGtkPrintSettings(void *p)
{
	return (GTK_PRINT_SETTINGS(p));
}

static GtkProgressBar *
toGtkProgressBar(void *p)
{
	return (GTK_PROGRESS_BAR(p));
}

static GtkLevelBar *
toGtkLevelBar(void *p)
{
	return (GTK_LEVEL_BAR(p));
}

static GtkBin *
toGtkBin(void *p)
{
	return (GTK_BIN(p));
}

static GtkWindow *
toGtkWindow(void *p)
{
	return (GTK_WINDOW(p));
}

static GtkBox *
toGtkBox(void *p)
{
	return (GTK_BOX(p));
}

static GtkStatusbar *
toGtkStatusbar(void *p)
{
	return (GTK_STATUSBAR(p));
}

static GtkLabel *
toGtkLabel(void *p)
{
	return (GTK_LABEL(p));
}

static GtkNotebook *
toGtkNotebook(void *p)
{
	return (GTK_NOTEBOOK(p));
}

static GtkEntry *
toGtkEntry(void *p)
{
	return (GTK_ENTRY(p));
}

static GtkEntryBuffer *
toGtkEntryBuffer(void *p)
{
	return (GTK_ENTRY_BUFFER(p));
}

static GtkEntryCompletion *
toGtkEntryCompletion(void *p)
{
	return (GTK_ENTRY_COMPLETION(p));
}

static GtkAdjustment *
toGtkAdjustment(void *p)
{
	return (GTK_ADJUSTMENT(p));
}

static GtkAccelGroup *
toGtkAccelGroup(void *p)
{
    return (GTK_ACCEL_GROUP(p));
}

static GtkAccelMap *
toGtkAccelMap(void *p)
{
    return (GTK_ACCEL_MAP(p));
}

static GtkTextTag *
toGtkTextTag(void *p)
{
	return (GTK_TEXT_TAG(p));
}

static GtkIconView *
toGtkIconView(void *p)
{
	return (GTK_ICON_VIEW(p));
}

static GtkImage *
toGtkImage(void *p)
{
	return (GTK_IMAGE(p));
}

static GtkButton *
toGtkButton(void *p)
{
	return (GTK_BUTTON(p));
}

static GtkScaleButton *
toGtkScaleButton(void *p)
{
	return (GTK_SCALE_BUTTON(p));
}

static GtkColorButton *
toGtkColorButton(void *p)
{
	return (GTK_COLOR_BUTTON(p));
}

static GtkViewport *
toGtkViewport(void *p)
{
	return (GTK_VIEWPORT(p));
}

static GtkVolumeButton *
toGtkVolumeButton(void *p)
{
	return (GTK_VOLUME_BUTTON(p));
}

static GtkScrollable *
toGtkScrollable(void *p)
{
	return (GTK_SCROLLABLE(p));
}

static GtkScrolledWindow *
toGtkScrolledWindow(void *p)
{
	return (GTK_SCROLLED_WINDOW(p));
}

static GtkMenuItem *
toGtkMenuItem(void *p)
{
	return (GTK_MENU_ITEM(p));
}

static GtkMenu *
toGtkMenu(void *p)
{
	return (GTK_MENU(p));
}

static GtkMenuShell *
toGtkMenuShell(void *p)
{
	return (GTK_MENU_SHELL(p));
}

static GtkMenuBar *
toGtkMenuBar(void *p)
{
	return (GTK_MENU_BAR(p));
}

static GtkSizeGroup *
toGtkSizeGroup(void *p)
{
	return (GTK_SIZE_GROUP(p));
}

static GtkSpinButton *
toGtkSpinButton(void *p)
{
	return (GTK_SPIN_BUTTON(p));
}

static GtkSpinner *
toGtkSpinner(void *p)
{
	return (GTK_SPINNER(p));
}

static GtkComboBox *
toGtkComboBox(void *p)
{
	return (GTK_COMBO_BOX(p));
}

static GtkComboBoxText *
toGtkComboBoxText(void *p)
{
	return (GTK_COMBO_BOX_TEXT(p));
}

static GtkLinkButton *
toGtkLinkButton(void *p)
{
	return (GTK_LINK_BUTTON(p));
}

static GtkLayout *
toGtkLayout(void *p)
{
	return (GTK_LAYOUT(p));
}

static GtkListStore *
toGtkListStore(void *p)
{
	return (GTK_LIST_STORE(p));
}

static GtkSwitch *
toGtkSwitch(void *p)
{
	return (GTK_SWITCH(p));
}

static GtkTextView *
toGtkTextView(void *p)
{
	return (GTK_TEXT_VIEW(p));
}

static GtkTextTagTable *
toGtkTextTagTable(void *p)
{
	return (GTK_TEXT_TAG_TABLE(p));
}

static GtkTextBuffer *
toGtkTextBuffer(void *p)
{
	return (GTK_TEXT_BUFFER(p));
}

static GtkTreeModel *
toGtkTreeModel(void *p)
{
	return (GTK_TREE_MODEL(p));
}

static GtkTreeModelFilter *
toGtkTreeModelFilter(void *p)
{
	return (GTK_TREE_MODEL_FILTER(p));
}

static GtkCellRenderer *
toGtkCellRenderer(void *p)
{
	return (GTK_CELL_RENDERER(p));
}

static GtkCellRendererPixbuf *
toGtkCellRendererPixbuf(void *p)
{
	return (GTK_CELL_RENDERER_PIXBUF(p));
}

static GtkCellRendererText *
toGtkCellRendererText(void *p)
{
	return (GTK_CELL_RENDERER_TEXT(p));
}

static GtkCellRendererToggle *
toGtkCellRendererToggle(void *p)
{
	return (GTK_CELL_RENDERER_TOGGLE(p));
}

static GtkCellLayout *
toGtkCellLayout(void *p)
{
	return (GTK_CELL_LAYOUT(p));
}

static GtkOrientable *
toGtkOrientable(void *p)
{
	return (GTK_ORIENTABLE(p));
}

static GtkTreeStore *
toGtkTreeStore (void *p)
{
	return (GTK_TREE_STORE(p));
}

static GtkTreeView *
toGtkTreeView(void *p)
{
	return (GTK_TREE_VIEW(p));
}

static GtkTreeViewColumn *
toGtkTreeViewColumn(void *p)
{
	return (GTK_TREE_VIEW_COLUMN(p));
}

static GtkTreeSelection *
toGtkTreeSelection(void *p)
{
	return (GTK_TREE_SELECTION(p));
}

static GtkTreeSortable *
toGtkTreeSortable(void *p)
{
	return (GTK_TREE_SORTABLE(p));
}

static GtkClipboard *
toGtkClipboard(void *p)
{
	return (GTK_CLIPBOARD(p));
}

static GtkDialog *
toGtkDialog(void *p)
{
	return (GTK_DIALOG(p));
}

static GtkMessageDialog *
toGtkMessageDialog(void *p)
{
	return (GTK_MESSAGE_DIALOG(p));
}

static GtkBuilder *
toGtkBuilder(void *p)
{
	return (GTK_BUILDER(p));
}

static GtkSeparatorMenuItem *
toGtkSeparatorMenuItem(void *p)
{
	return (GTK_SEPARATOR_MENU_ITEM(p));
}

static GtkCheckButton *
toGtkCheckButton(void *p)
{
	return (GTK_CHECK_BUTTON(p));
}

static GtkToggleButton *
toGtkToggleButton(void *p)
{
	return (GTK_TOGGLE_BUTTON(p));
}

static GtkFontButton *
toGtkFontButton(void *p)
{
	return (GTK_FONT_BUTTON(p));
}

static GtkFrame *
toGtkFrame(void *p)
{
	return (GTK_FRAME(p));
}

static GtkAspectFrame *
toGtkAspectFrame(void *p)
{
	return (GTK_ASPECT_FRAME(p));
}

static GtkSeparator *
toGtkSeparator(void *p)
{
	return (GTK_SEPARATOR(p));
}

static GtkScale*
toGtkScale(void *p)
{
	return (GTK_SCALE(p));
}

static GtkScrollbar *
toGtkScrollbar(void *p)
{
	return (GTK_SCROLLBAR(p));
}

static GtkRange *
toGtkRange(void *p)
{
	return (GTK_RANGE(p));
}

static GtkSearchEntry *
toGtkSearchEntry(void *p)
{
	return (GTK_SEARCH_ENTRY(p));
}

static GtkOffscreenWindow *
toGtkOffscreenWindow(void *p)
{
	return (GTK_OFFSCREEN_WINDOW(p));
}

static GtkExpander *
toGtkExpander(void *p)
{
	return (GTK_EXPANDER(p));
}

static GtkFileChooser *
toGtkFileChooser(void *p)
{
	return (GTK_FILE_CHOOSER(p));
}

static GtkFileChooserButton *
toGtkFileChooserButton(void *p)
{
	return (GTK_FILE_CHOOSER_BUTTON(p));
}

static GtkFileChooserDialog *
toGtkFileChooserDialog(void *p)
{
	return (GTK_FILE_CHOOSER_DIALOG(p));
}

static GtkFileChooserWidget *
toGtkFileChooserWidget(void *p)
{
	return (GTK_FILE_CHOOSER_WIDGET(p));
}

static GtkFileFilter *
toGtkFileFilter(void *p)
{
	return (GTK_FILE_FILTER(p));
}

static GtkMenuButton *
toGtkMenuButton(void *p)
{
	return (GTK_MENU_BUTTON(p));
}

static GtkRadioButton *
toGtkRadioButton(void *p)
{
	return (GTK_RADIO_BUTTON(p));
}

static GtkRecentChooser *
toGtkRecentChooser(void *p)
{
	return (GTK_RECENT_CHOOSER(p));
}

static GtkRecentChooserMenu *
toGtkRecentChooserMenu(void *p)
{
	return (GTK_RECENT_CHOOSER_MENU(p));
}

static GtkColorChooser *
toGtkColorChooser(void *p)
{
	return (GTK_COLOR_CHOOSER(p));
}

static GtkRecentFilter *
toGtkRecentFilter(void *p)
{
	return (GTK_RECENT_FILTER(p));
}

static GtkRecentManager *
toGtkRecentManager(void *p)
{
	return (GTK_RECENT_MANAGER(p));
}

static GtkCheckMenuItem *
toGtkCheckMenuItem(void *p)
{
	return (GTK_CHECK_MENU_ITEM(p));
}

static GtkRadioMenuItem *
toGtkRadioMenuItem(void *p)
{
	return (GTK_RADIO_MENU_ITEM(p));
}

static GtkToolItem *
toGtkToolItem(void *p)
{
	return (GTK_TOOL_ITEM(p));
}

static GtkToolbar *
toGtkToolbar(void *p)
{
	return (GTK_TOOLBAR(p));
}

static GtkTooltip *
toGtkTooltip(void *p)
{
	return (GTK_TOOLTIP(p));
}

static GtkEditable *
toGtkEditable(void *p)
{
	return (GTK_EDITABLE(p));
}

static GtkToolButton *
toGtkToolButton(void *p)
{
	return (GTK_TOOL_BUTTON(p));
}

static GtkSeparatorToolItem *
toGtkSeparatorToolItem(void *p)
{
	return (GTK_SEPARATOR_TOOL_ITEM(p));
}

static GtkCssProvider *
toGtkCssProvider(void *p)
{
        return (GTK_CSS_PROVIDER(p));
}

static GtkStyleContext *
toGtkStyleContext(void *p)
{
        return (GTK_STYLE_CONTEXT(p));
}

static GtkStyleProvider *
toGtkStyleProvider(void *p)
{
        return (GTK_STYLE_PROVIDER(p));
}

static GtkInfoBar *
toGtkInfoBar(void *p)
{
	return (GTK_INFO_BAR(p));
}

static GMenuModel *
toGMenuModel(void *p)
{
	return (G_MENU_MODEL(p));
}

static GActionGroup *
toGActionGroup(void *p)
{
	return (G_ACTION_GROUP(p));
}

static GType *
alloc_types(int n) {
	return ((GType *)g_new0(GType, n));
}

static void
set_type(GType *types, int n, GType t)
{
	types[n] = t;
}

static GtkTreeViewColumn *
_gtk_tree_view_column_new_with_attributes_one(const gchar *title,
    GtkCellRenderer *renderer, const gchar *attribute, gint column)
{
	GtkTreeViewColumn	*tvc;

	tvc = gtk_tree_view_column_new_with_attributes(title, renderer,
	    attribute, column, NULL);
	return (tvc);
}

static void
_gtk_list_store_set(GtkListStore *list_store, GtkTreeIter *iter, gint column,
	void* value)
{
	gtk_list_store_set(list_store, iter, column, value, -1);
}

static void
_gtk_tree_store_set(GtkTreeStore *store, GtkTreeIter *iter, gint column,
	void* value)
{
	gtk_tree_store_set(store, iter, column, value, -1);
}

extern gboolean substring_match_equal_func(GtkTreeModel *model,
                                          gint column,
                                          gchar *key,
                                          GtkTreeIter *iter,
                                          gpointer data);

static GtkWidget *
_gtk_message_dialog_new(GtkWindow *parent, GtkDialogFlags flags,
    GtkMessageType type, GtkButtonsType buttons, char *msg)
{
	GtkWidget		*w;

	w = gtk_message_dialog_new(parent, flags, type, buttons, "%s", msg);
	return (w);
}

static GtkWidget *
_gtk_message_dialog_new_with_markup(GtkWindow *parent, GtkDialogFlags flags,
    GtkMessageType type, GtkButtonsType buttons, char *msg)
{
	GtkWidget		*w;

	w = gtk_message_dialog_new_with_markup(parent, flags, type, buttons,
	    "%s", msg);
	return (w);
}

static void
_gtk_message_dialog_format_secondary_text(GtkMessageDialog *message_dialog,
    const gchar *msg)
{
	gtk_message_dialog_format_secondary_text(message_dialog, "%s", msg);
}

static void
_gtk_message_dialog_format_secondary_markup(GtkMessageDialog *message_dialog,
    const gchar *msg)
{
	gtk_message_dialog_format_secondary_markup(message_dialog, "%s", msg);
}

static const gchar *
object_get_class_name(GObject *object)
{
	return G_OBJECT_CLASS_NAME(G_OBJECT_GET_CLASS(object));
}

static GtkWidget *
gtk_file_chooser_dialog_new_1(
	const gchar *title,
	GtkWindow *parent,
	GtkFileChooserAction action,
	const gchar *first_button_text, int first_button_id
) {
	return gtk_file_chooser_dialog_new(
		title, parent, action,
		first_button_text, first_button_id,
		NULL);
}

static GtkWidget *
gtk_file_chooser_dialog_new_2(
	const gchar *title,
	GtkWindow *parent,
	GtkFileChooserAction action,
	const gchar *first_button_text, int first_button_id,
	const gchar *second_button_text, int second_button_id
) {
	return gtk_file_chooser_dialog_new(
		title, parent, action,
		first_button_text, first_button_id,
		second_button_text, second_button_id,
		NULL);
}

static void _gtk_widget_hide_on_delete(GtkWidget* w) {
	g_signal_connect(GTK_WIDGET(w), "delete-event", G_CALLBACK(gtk_widget_hide_on_delete), NULL);
}

static inline gchar** make_strings(int count) {
	return (gchar**)malloc(sizeof(gchar*) * count);
}

static inline void destroy_strings(gchar** strings) {
	free(strings);
}

static inline gchar* get_string(gchar** strings, int n) {
	return strings[n];
}

static inline void set_string(gchar** strings, int n, gchar* str) {
	strings[n] = str;
}

static inline gchar** next_gcharptr(gchar** s) { return (s+1); }

extern void goBuilderConnect (GtkBuilder *builder,
                          GObject *object,
                          gchar *signal_name,
                          gchar *handler_name,
                          GObject *connect_object,
                          GConnectFlags flags,
                          gpointer user_data);

static inline void _gtk_builder_connect_signals_full(GtkBuilder *builder) {
	gtk_builder_connect_signals_full(builder, (GtkBuilderConnectFunc)(goBuilderConnect), NULL);
}

extern void goPrintSettings (gchar *key,
	                     gchar *value,
                         gpointer user_data);

static inline void _gtk_print_settings_foreach(GtkPrintSettings *ps, gpointer user_data) {
	gtk_print_settings_foreach(ps, (GtkPrintSettingsFunc)(goPrintSettings), user_data);
}

extern void goPageSetupDone (GtkPageSetup *setup,
                         gpointer data);

static inline void _gtk_print_run_page_setup_dialog_async(GtkWindow *parent, GtkPageSetup *setup,
	GtkPrintSettings *settings, gpointer data) {
	gtk_print_run_page_setup_dialog_async(parent, setup, settings,
		(GtkPageSetupDoneFunc)(goPageSetupDone), data);
}
