//go:build darwin

package display

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework IOKit -framework CoreFoundation

#include <CoreGraphics/CoreGraphics.h>
#include <IOKit/graphics/IOGraphicsLib.h>
#include <stdlib.h>
#include <string.h>

// getDisplayName retrieves the product name of a display from IOKit.
// Returns a malloc'd string that must be freed by the caller.
char* getDisplayName(CGDirectDisplayID displayID) {
    io_service_t service = CGDisplayIOServicePort(displayID);
    if (service == 0) {
        return strdup("Unknown");
    }

    CFDictionaryRef info = IODisplayCreateInfoDictionary(service, kIODisplayOnlyPreferredName);
    if (info == NULL) {
        return strdup("Unknown");
    }

    CFDictionaryRef names = CFDictionaryGetValue(info, CFSTR(kDisplayProductName));
    if (names == NULL || CFGetTypeID(names) != CFDictionaryGetTypeID()) {
        CFRelease(info);
        return strdup("Unknown");
    }

    CFIndex count = CFDictionaryGetCount(names);
    if (count == 0) {
        CFRelease(info);
        return strdup("Unknown");
    }

    // Get the first (and usually only) localized name
    const void *values[1];
    CFDictionaryGetKeysAndValues(names, NULL, values);
    CFStringRef name = (CFStringRef)values[0];

    char *result;
    if (name != NULL && CFGetTypeID(name) == CFStringGetTypeID()) {
        CFIndex length = CFStringGetLength(name);
        CFIndex maxSize = CFStringGetMaximumSizeForEncoding(length, kCFStringEncodingUTF8) + 1;
        result = malloc(maxSize);
        if (result != NULL) {
            if (!CFStringGetCString(name, result, maxSize, kCFStringEncodingUTF8)) {
                free(result);
                result = strdup("Unknown");
            }
        } else {
            result = strdup("Unknown");
        }
    } else {
        result = strdup("Unknown");
    }

    CFRelease(info);
    return result;
}

// DisplayData holds display information returned from C.
typedef struct {
    CGDirectDisplayID id;
    long width;
    int isMain;
    char *name;
} DisplayData;

// getDisplays populates an array with display information.
// Returns the number of displays found, or -1 on error.
int getDisplays(DisplayData *displays, int maxDisplays) {
    CGDirectDisplayID displayIDs[16];
    uint32_t displayCount;

    CGError err = CGGetActiveDisplayList(16, displayIDs, &displayCount);
    if (err != kCGErrorSuccess) {
        return -1;
    }

    CGDirectDisplayID mainID = CGMainDisplayID();

    int count = (displayCount < maxDisplays) ? displayCount : maxDisplays;
    for (int i = 0; i < count; i++) {
        displays[i].id = displayIDs[i];
        displays[i].width = CGDisplayPixelsWide(displayIDs[i]);
        displays[i].isMain = (displayIDs[i] == mainID) ? 1 : 0;
        displays[i].name = getDisplayName(displayIDs[i]);
    }

    return count;
}

// freeDisplayData frees the memory allocated for display names.
void freeDisplayData(DisplayData *displays, int count) {
    for (int i = 0; i < count; i++) {
        if (displays[i].name != NULL) {
            free(displays[i].name);
        }
    }
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

// Enumerate returns information about all active displays.
func Enumerate() ([]Info, error) {
	var displays [16]C.DisplayData

	count := C.getDisplays(&displays[0], 16)
	if count < 0 {
		return nil, errors.New("failed to enumerate displays")
	}
	defer C.freeDisplayData(&displays[0], count)

	result := make([]Info, count)
	for i := 0; i < int(count); i++ {
		result[i] = Info{
			ID:    uint32(displays[i].id),
			Name:  C.GoString(displays[i].name),
			Width: int64(displays[i].width),
			Main:  displays[i].isMain != 0,
		}
	}

	return result, nil
}

// MainWidth returns the width of the main display in pixels.
func MainWidth() (int64, error) {
	width := C.CGDisplayPixelsWide(C.CGMainDisplayID())
	if width == 0 {
		return 0, errors.New("failed to get main display width")
	}
	return int64(width), nil
}

// available indicates whether display detection is available on this platform.
const available = true

// Available returns true if display detection is supported on this platform.
func Available() bool {
	return available
}

// dummyUnsafe is used to ensure the unsafe package is imported for CGO.
var _ = unsafe.Pointer(nil)
