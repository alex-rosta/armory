(function () {
  "use strict";

  function store(key, value) {
    try {
      localStorage.setItem(key, value);
    } catch (e) {}
  }

  function read(key) {
    try {
      return localStorage.getItem(key);
    } catch (e) {
      return null;
    }
  }

  // Theme toggle
  var toggle = document.getElementById("theme-toggle");
  if (toggle) {
    var root = document.documentElement;
    toggle.setAttribute("aria-pressed", String(root.classList.contains("dark")));
    toggle.addEventListener("click", function () {
      var dark = root.classList.toggle("dark");
      toggle.setAttribute("aria-pressed", String(dark));
      store("theme", dark ? "dark" : "light");
    });
  }

  // Remember the last region unless the URL already specifies one
  var region = document.querySelector("[data-region]");
  if (region) {
    var saved = read("region");
    if (saved && !new URLSearchParams(location.search).has("region")) {
      region.value = saved;
      if (region.value !== saved) region.selectedIndex = 0;
    }
    region.addEventListener("change", function () {
      store("region", region.value);
    });
  }

  // Show progress while a search is running
  var form = document.querySelector("[data-search-form]");
  if (form) {
    var button = form.querySelector("[data-submit]");
    var icon = form.querySelector("[data-submit-icon]");
    var label = form.querySelector("[data-submit-label]");
    var idleLabel = label.textContent;
    var idleIcon = icon.className;

    var reset = function () {
      button.disabled = false;
      button.removeAttribute("aria-busy");
      label.textContent = idleLabel;
      icon.className = idleIcon;
    };

    form.addEventListener("submit", function () {
      if (!form.checkValidity()) return;
      // Disable on the next tick so the submit itself is not cancelled
      setTimeout(function () {
        button.disabled = true;
        button.setAttribute("aria-busy", "true");
        label.textContent = "Searching…";
        icon.className = "bi bi-arrow-repeat animate-spin";
      }, 0);
    });

    // Restore the button when coming back through the browser cache
    window.addEventListener("pageshow", reset);
  }
})();
