#import <Cocoa/Cocoa.h>
#import <ApplicationServices/ApplicationServices.h>
#import <dispatch/dispatch.h>
#import <unistd.h>

static volatile int fastswitchHotKeyPressCount = 0;
static volatile int fastswitchCommandKeyReleaseCount = 0;
static BOOL fastswitchCommandKeyDown = NO;
static CFMachPortRef fastswitchEventTap = nil;
static CFRunLoopSourceRef fastswitchEventTapSource = nil;
static NSStatusItem *fastswitchStatusItem = nil;
static volatile int fastswitchStatusAction = 0;
static NSMutableDictionary<NSString *, NSString *> *fastswitchIconCache = nil;
static const CGFloat fastswitchWindowCornerRadius = 18.0;

@interface FastSwitchStatusTarget : NSObject
@end

@implementation FastSwitchStatusTarget
- (void)openSwitcher:(id)sender {
  fastswitchStatusAction = 1;
}

- (void)refreshWindows:(id)sender {
  fastswitchStatusAction = 2;
}

- (void)quitApp:(id)sender {
  fastswitchStatusAction = 3;
}
@end

static FastSwitchStatusTarget *fastswitchStatusTarget = nil;

static BOOL fastswitch_canUseString(id value) {
  return value != nil && value != [NSNull null] && [value isKindOfClass:[NSString class]];
}

static NSString *fastswitch_icon_base64_for_pid(NSNumber *pidNumber, NSMutableDictionary *iconCache) {
  if (iconCache == nil) {
    return @"";
  }
  NSString *cacheKey = pidNumber.stringValue;
  NSString *cached = iconCache[cacheKey];
  if (cached != nil) {
    return cached;
  }

  NSRunningApplication *app = [NSRunningApplication runningApplicationWithProcessIdentifier:(pid_t)pidNumber.intValue];
  if (app == nil || app.icon == nil) {
    iconCache[cacheKey] = @"";
    return @"";
  }

  NSImage *icon = app.icon.copy;
  icon.size = NSMakeSize(128, 128);
  NSData *tiffData = [icon TIFFRepresentation];
  if (tiffData == nil) {
    iconCache[cacheKey] = @"";
    return @"";
  }

  NSBitmapImageRep *rep = [NSBitmapImageRep imageRepWithData:tiffData];
  NSData *pngData = [rep representationUsingType:NSBitmapImageFileTypePNG properties:@{}];
  if (pngData == nil) {
    iconCache[cacheKey] = @"";
    return @"";
  }

  NSString *base64 = [pngData base64EncodedStringWithOptions:0];
  iconCache[cacheKey] = base64;
  return base64;
}


static void fastswitch_apply_corner_mask(NSView *view, CGFloat cornerRadius) {
  if (view == nil) {
    return;
  }

  view.wantsLayer = YES;
  view.layer.cornerRadius = cornerRadius;
  view.layer.masksToBounds = YES;
  view.layer.backgroundColor = NSColor.clearColor.CGColor;
}

static void fastswitch_apply_window_appearance(NSWindow *window) {
  if (window == nil) {
    return;
  }

  window.opaque = NO;
  window.backgroundColor = NSColor.clearColor;
  window.hasShadow = YES;

  NSView *contentView = window.contentView;
  if (contentView == nil) {
    return;
  }

  fastswitch_apply_corner_mask(contentView, fastswitchWindowCornerRadius);

  NSView *frameView = contentView.superview;
  if (frameView != nil) {
    fastswitch_apply_corner_mask(frameView, fastswitchWindowCornerRadius);
  }
}

void fastswitch_configure_window_appearance(void) {
  dispatch_async(dispatch_get_main_queue(), ^{
    NSWindow *window = NSApp.keyWindow;
    if (window == nil && NSApp.windows.count > 0) {
      window = NSApp.windows.firstObject;
    }
    fastswitch_apply_window_appearance(window);
  });
}

static CGEventRef fastswitch_event_tap_callback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *userInfo) {
  if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
    if (fastswitchEventTap != nil) {
      CGEventTapEnable(fastswitchEventTap, true);
    }
    return event;
  }

  if (type == kCGEventFlagsChanged) {
    CGEventFlags flags = CGEventGetFlags(event);
    BOOL commandPressed = (flags & kCGEventFlagMaskCommand) == kCGEventFlagMaskCommand;
    if (fastswitchCommandKeyDown && !commandPressed) {
      fastswitchCommandKeyReleaseCount += 1;
      fastswitchCommandKeyDown = NO;
    } else if (commandPressed) {
      fastswitchCommandKeyDown = YES;
    }
    return event;
  }

  if (type == kCGEventKeyDown) {
    CGEventFlags flags = CGEventGetFlags(event);
    CGKeyCode keycode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
    BOOL commandPressed = (flags & kCGEventFlagMaskCommand) == kCGEventFlagMaskCommand;

    if (commandPressed && keycode == (CGKeyCode)48) {
      fastswitchHotKeyPressCount += 1;
      fastswitchCommandKeyDown = YES;
      return NULL;
    }
  }

  return event;
}

char *fastswitch_copy_windows_json(void) {
  CFArrayRef windowList = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly | kCGWindowListExcludeDesktopElements, kCGNullWindowID);
  if (windowList == nil) {
    return strdup("[]");
  }

  pid_t currentPID = getpid();
  if (fastswitchIconCache == nil) {
    fastswitchIconCache = [NSMutableDictionary dictionary];
  }
  NSMutableArray *windows = [NSMutableArray array];
  NSArray *entries = CFBridgingRelease(windowList);

  for (NSDictionary *entry in entries) {
    NSNumber *layerNumber = entry[(id)kCGWindowLayer];
    NSNumber *alphaNumber = entry[(id)kCGWindowAlpha];
    NSString *ownerName = entry[(id)kCGWindowOwnerName];
    NSDictionary *bounds = entry[(id)kCGWindowBounds];
    NSNumber *pidNumber = entry[(id)kCGWindowOwnerPID];

    if (![layerNumber isKindOfClass:[NSNumber class]] || layerNumber.intValue != 0) {
      continue;
    }
    if ([alphaNumber isKindOfClass:[NSNumber class]] && alphaNumber.doubleValue <= 0.0) {
      continue;
    }
    if (![ownerName isKindOfClass:[NSString class]] || ownerName.length == 0) {
      continue;
    }
    if (![pidNumber isKindOfClass:[NSNumber class]] || ![bounds isKindOfClass:[NSDictionary class]]) {
      continue;
    }
    if (pidNumber.intValue == currentPID) {
      continue;
    }

    double width = [bounds[@"Width"] doubleValue];
    double height = [bounds[@"Height"] doubleValue];
    if (width < 32.0 || height < 32.0) {
      continue;
    }

    NSString *title = @"";
    id rawTitle = entry[(id)kCGWindowName];
    if (fastswitch_canUseString(rawTitle)) {
      title = rawTitle;
    }

    [windows addObject:@{
      @"ownerName": ownerName,
      @"title": title,
      @"icon": fastswitch_icon_base64_for_pid(pidNumber, fastswitchIconCache),
      @"pid": pidNumber,
      @"layer": layerNumber,
      @"x": @([bounds[@"X"] doubleValue]),
      @"y": @([bounds[@"Y"] doubleValue]),
      @"width": @(width),
      @"height": @(height)
    }];
  }

  NSData *jsonData = [NSJSONSerialization dataWithJSONObject:windows options:0 error:nil];
  if (jsonData == nil) {
    return strdup("[]");
  }

  NSString *json = [[NSString alloc] initWithData:jsonData encoding:NSUTF8StringEncoding];
  return strdup(json.UTF8String);
}

bool fastswitch_accessibility_trusted(void) {
  return AXIsProcessTrusted();
}

bool fastswitch_request_accessibility(void) {
  NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
  return AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
}

bool fastswitch_screen_recording_granted(void) {
  if (@available(macOS 10.15, *)) {
    return CGPreflightScreenCaptureAccess();
  }
  return true;
}

bool fastswitch_request_screen_recording(void) {
  if (@available(macOS 10.15, *)) {
    return CGRequestScreenCaptureAccess();
  }
  return true;
}

bool fastswitch_activate_app(int pid) {
  NSRunningApplication *app = [NSRunningApplication runningApplicationWithProcessIdentifier:(pid_t)pid];
  if (app == nil) {
    return false;
  }

  return [app activateWithOptions:0];
}

bool fastswitch_activate_window(int pid, const char *title) {
  NSRunningApplication *app = [NSRunningApplication runningApplicationWithProcessIdentifier:(pid_t)pid];
  if (app == nil) {
    return false;
  }

  [app activateWithOptions:0];

  AXUIElementRef applicationElement = AXUIElementCreateApplication((pid_t)pid);
  if (applicationElement == nil) {
    return false;
  }

  CFArrayRef windowsRef = nil;
  AXError windowsError = AXUIElementCopyAttributeValue(applicationElement, kAXWindowsAttribute, (CFTypeRef *)&windowsRef);
  if (windowsError != kAXErrorSuccess || windowsRef == nil) {
    CFRelease(applicationElement);
    return false;
  }

  NSString *desiredTitle = title != NULL ? [NSString stringWithUTF8String:title] : @"";
  NSArray *windows = CFBridgingRelease(windowsRef);
  BOOL activated = NO;

  for (id entry in windows) {
    AXUIElementRef window = (__bridge AXUIElementRef)entry;
    CFTypeRef titleValue = nil;
    NSString *windowTitle = @"";

    if (AXUIElementCopyAttributeValue(window, kAXTitleAttribute, &titleValue) == kAXErrorSuccess && titleValue != nil) {
      if (CFGetTypeID(titleValue) == CFStringGetTypeID()) {
        windowTitle = CFBridgingRelease(titleValue);
      } else {
        CFRelease(titleValue);
      }
    }

    if (desiredTitle.length > 0 && ![windowTitle isEqualToString:desiredTitle]) {
      continue;
    }

    AXUIElementPerformAction(window, kAXRaiseAction);
    AXUIElementSetAttributeValue(window, kAXMainAttribute, kCFBooleanTrue);
    AXUIElementSetAttributeValue(window, kAXFocusedAttribute, kCFBooleanTrue);
    activated = YES;
    break;
  }

  CFRelease(applicationElement);
  return activated;
}

int fastswitch_frontmost_app_pid(void) {
  NSRunningApplication *app = [[NSWorkspace sharedWorkspace] frontmostApplication];
  if (app == nil) {
    return 0;
  }
  return (int)app.processIdentifier;
}

bool fastswitch_register_option_tab_hotkey(void) {
  __block BOOL created = YES;

  dispatch_sync(dispatch_get_main_queue(), ^{
    if (fastswitchEventTap != nil) {
      return;
    }

    CGEventMask mask = CGEventMaskBit(kCGEventKeyDown) | CGEventMaskBit(kCGEventFlagsChanged);
    fastswitchEventTap = CGEventTapCreate(
      kCGSessionEventTap,
      kCGHeadInsertEventTap,
      kCGEventTapOptionDefault,
      mask,
      fastswitch_event_tap_callback,
      nil
    );

    if (fastswitchEventTap == nil) {
      created = NO;
      return;
    }

    fastswitchEventTapSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, fastswitchEventTap, 0);
    if (fastswitchEventTapSource == nil) {
      CFRelease(fastswitchEventTap);
      fastswitchEventTap = nil;
      created = NO;
      return;
    }

    CFRunLoopAddSource(CFRunLoopGetMain(), fastswitchEventTapSource, kCFRunLoopCommonModes);
    CGEventTapEnable(fastswitchEventTap, true);
  });

  return created;
}

int fastswitch_consume_option_tab_press_count(void) {
  int count = fastswitchHotKeyPressCount;
  fastswitchHotKeyPressCount = 0;
  return count;
}

int fastswitch_consume_option_key_release_count(void) {
  int count = fastswitchCommandKeyReleaseCount;
  fastswitchCommandKeyReleaseCount = 0;
  return count;
}

bool fastswitch_register_status_item(void) {
  __block BOOL created = YES;

  dispatch_sync(dispatch_get_main_queue(), ^{
    if (fastswitchStatusItem != nil) {
      return;
    }

    fastswitchStatusTarget = [FastSwitchStatusTarget new];
    fastswitchStatusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
    if (fastswitchStatusItem == nil) {
      created = NO;
      return;
    }

    fastswitchStatusItem.button.title = @"FS";
    fastswitchStatusItem.button.toolTip = @"Fast Switch";

    NSMenu *menu = [[NSMenu alloc] initWithTitle:@"Fast Switch"];

    NSMenuItem *openItem = [[NSMenuItem alloc] initWithTitle:@"Open Switcher" action:@selector(openSwitcher:) keyEquivalent:@""];
    openItem.target = fastswitchStatusTarget;
    [menu addItem:openItem];

    NSMenuItem *refreshItem = [[NSMenuItem alloc] initWithTitle:@"Refresh Windows" action:@selector(refreshWindows:) keyEquivalent:@""];
    refreshItem.target = fastswitchStatusTarget;
    [menu addItem:refreshItem];

    [menu addItem:[NSMenuItem separatorItem]];

    NSMenuItem *quitItem = [[NSMenuItem alloc] initWithTitle:@"Quit Fast Switch" action:@selector(quitApp:) keyEquivalent:@""];
    quitItem.target = fastswitchStatusTarget;
    [menu addItem:quitItem];

    fastswitchStatusItem.menu = menu;
  });

  return created;
}

int fastswitch_consume_status_action(void) {
  int action = fastswitchStatusAction;
  fastswitchStatusAction = 0;
  return action;
}

void fastswitch_free_string(char *value) {
  if (value != nil) {
    free(value);
  }
}
