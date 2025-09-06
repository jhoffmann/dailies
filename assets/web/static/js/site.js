import htmx from "htmx.org";
import "htmx-ext-json-enc";
import "htmx-ext-ws";
import _hyperscript from "hyperscript.org";

import "./lib/notifications.js";

// Expose htmx globally
window.htmx = htmx;

_hyperscript.browserInit();
