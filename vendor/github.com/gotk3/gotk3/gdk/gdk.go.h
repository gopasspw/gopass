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

#include <stdlib.h>

// Type Casting
static GdkAtom
toGdkAtom(void *p)
{
	return ((GdkAtom)p);
}

static GdkDevice *
toGdkDevice(void *p)
{
	return (GDK_DEVICE(p));
}

static GdkCursor *
toGdkCursor(void *p)
{
	return (GDK_CURSOR(p));
}

static GdkDeviceManager *
toGdkDeviceManager(void *p)
{
	return (GDK_DEVICE_MANAGER(p));
}

static GdkDisplay *
toGdkDisplay(void *p)
{
	return (GDK_DISPLAY(p));
}

static GdkDragContext *
toGdkDragContext(void *p)
{
	return (GDK_DRAG_CONTEXT(p));
}

static GdkPixbuf *
toGdkPixbuf(void *p)
{
	return (GDK_PIXBUF(p));
}

static gboolean
_gdk_pixbuf_save_png(GdkPixbuf *pixbuf,
const char *filename, GError ** err, const char *compression)
{
	return gdk_pixbuf_save(pixbuf, filename, "png", err, "compression", compression, NULL);
}

static gboolean
_gdk_pixbuf_save_jpeg(GdkPixbuf *pixbuf,
const char *filename, GError ** err, const char *quality)
{
	return gdk_pixbuf_save(pixbuf, filename, "jpeg", err, "quality", quality, NULL);
}

static GdkPixbufLoader *
toGdkPixbufLoader(void *p)
{
	return (GDK_PIXBUF_LOADER(p));
}

static GdkScreen *
toGdkScreen(void *p)
{
	return (GDK_SCREEN(p));
}

static GdkVisual *
toGdkVisual(void *p)
{
	return (GDK_VISUAL(p));
}

static GdkWindow *
toGdkWindow(void *p)
{
	return (GDK_WINDOW(p));
}
