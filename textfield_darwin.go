// 16 july 2014

package ui

import (
	"unsafe"
)

// #include "objc_darwin.h"
import "C"

type textfield struct {
	_id     C.id
	changed *event
	invalid C.id
}

func finishNewTextField(id C.id) *textfield {
	t := &textfield{
		_id:     id,
		changed: newEvent(),
	}
	C.textfieldSetDelegate(t._id, unsafe.Pointer(t))
	return t
}

func newTextField() *textfield {
	return finishNewTextField(C.newTextField())
}

func newPasswordField() *textfield {
	return finishNewTextField(C.newPasswordField())
}

func (t *textfield) Text() string {
	return C.GoString(C.textfieldText(t._id))
}

func (t *textfield) SetText(text string) {
	ctext := C.CString(text)
	defer C.free(unsafe.Pointer(ctext))
	C.textfieldSetText(t._id, ctext)
}

func (t *textfield) OnChanged(f func()) {
	t.changed.set(f)
}

func (t *textfield) Invalid(reason string) {
	if t.invalid != nil {
		C.textfieldCloseInvalidPopover(t.invalid)
		t.invalid = nil
	}
	if reason == "" {
		return
	}
	creason := C.CString(reason)
	defer C.free(unsafe.Pointer(creason))
	t.invalid = C.textfieldOpenInvalidPopover(t._id, creason)
}

//export textfieldChanged
func textfieldChanged(data unsafe.Pointer) {
	t := (*textfield)(data)
	t.changed.fire()
}

func (t *textfield) id() C.id {
	return t._id
}

func (t *textfield) setParent(p *controlParent) {
	basesetParent(t, p)
}

func (t *textfield) allocate(x int, y int, width int, height int, d *sizing) []*allocation {
	return baseallocate(t, x, y, width, height, d)
}

func (t *textfield) preferredSize(d *sizing) (width, height int) {
	_, height = basepreferredSize(t, d)
	// the returned width is based on the contents; use this instead
	return C.textfieldWidth, height
}

func (t *textfield) commitResize(a *allocation, d *sizing) {
	basecommitResize(t, a, d)
}

func (t *textfield) getAuxResizeInfo(d *sizing) {
	basegetAuxResizeInfo(t, d)
}
