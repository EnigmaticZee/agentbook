// Agentbook site: theme toggle, reveal on scroll, tabs, copy buttons and the terminal demo.
(function () {
  "use strict";
  var root = document.documentElement;
  var reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  // ---- Theme ----
  function isDark() {
    var t = root.dataset.theme;
    if (t) return t === "dark";
    return window.matchMedia("(prefers-color-scheme: dark)").matches;
  }
  document.getElementById("theme").addEventListener("click", function () {
    var next = isDark() ? "light" : "dark";
    root.dataset.theme = next;
    try { localStorage.setItem("ab-theme", next); } catch (e) {}
  });

  // ---- Reveal on scroll ----
  var items = document.querySelectorAll("[data-reveal]");
  if (reduced || !("IntersectionObserver" in window)) {
    items.forEach(function (el) { el.classList.add("in"); });
  } else {
    var io = new IntersectionObserver(function (entries) {
      entries.forEach(function (e) {
        if (!e.isIntersecting) return;
        var el = e.target, sibs = Array.prototype.indexOf.call(el.parentNode.children, el);
        el.style.transitionDelay = Math.min(sibs, 5) * 80 + "ms";
        el.classList.add("in");
        io.unobserve(el);
      });
    }, { rootMargin: "0px 0px -6% 0px", threshold: 0.04 });
    items.forEach(function (el) { io.observe(el); });
  }

  // ---- Contents: mark the chapter being read ----
  var tocLinks = document.querySelectorAll(".toc a");
  if ("IntersectionObserver" in window && tocLinks.length) {
    var byId = {};
    tocLinks.forEach(function (a) { byId[a.getAttribute("href").slice(1)] = a; });
    var tocIO = new IntersectionObserver(function (entries) {
      entries.forEach(function (e) {
        if (!e.isIntersecting) return;
        tocLinks.forEach(function (a) { a.classList.remove("on"); });
        var a = byId[e.target.id];
        if (a) a.classList.add("on");
      });
    }, { rootMargin: "-30% 0px -60% 0px" });
    document.querySelectorAll(".chapter").forEach(function (c) { tocIO.observe(c); });
    // The last chapter is short: mark it once the page is scrolled to the end.
    window.addEventListener("scroll", function () {
      if (window.innerHeight + window.scrollY >= document.documentElement.scrollHeight - 4) {
        tocLinks.forEach(function (a) { a.classList.remove("on"); });
        tocLinks[tocLinks.length - 1].classList.add("on");
      }
    }, { passive: true });
  }

  // ---- Tabs (install) ----
  function tabs(list, attr, onChange) {
    var buttons = list.querySelectorAll("button");
    function select(btn, focus) {
      buttons.forEach(function (b) {
        var on = b === btn;
        b.setAttribute("aria-selected", on ? "true" : "false");
        b.tabIndex = on ? 0 : -1;
      });
      if (focus) btn.focus();
      onChange(btn.getAttribute(attr));
    }
    buttons.forEach(function (b, i) {
      b.tabIndex = b.getAttribute("aria-selected") === "true" ? 0 : -1;
      b.addEventListener("click", function () { select(b); });
      b.addEventListener("keydown", function (e) {
        var d = e.key === "ArrowRight" ? 1 : e.key === "ArrowLeft" ? -1 : 0;
        if (!d) return;
        e.preventDefault();
        select(buttons[(i + d + buttons.length) % buttons.length], true);
      });
    });
  }
  tabs(document.querySelector(".inst-tabs"), "data-tab", function (name) {
    document.querySelectorAll(".inst .pane").forEach(function (p) { p.hidden = p.dataset.pane !== name; });
  });

  // ---- Copy ----
  document.querySelectorAll(".copy").forEach(function (btn) {
    btn.addEventListener("click", function () {
      var text = btn.parentNode.querySelector("code").textContent;
      var done = function () {
        btn.textContent = "Copied"; btn.classList.add("done");
        setTimeout(function () { btn.textContent = "Copy"; btn.classList.remove("done"); }, 1600);
      };
      if (navigator.clipboard) navigator.clipboard.writeText(text).then(done, function () {});
    });
  });

  // ---- Terminal demo ----
  // Each demo is a list of steps: ["type", text], ["out", html], ["screen", html], ["wait", ms].
  // The output mirrors what the real CLI prints.
  function esc(s) { return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;"); }
  function s(cls, text) { return '<span class="' + cls + '">' + esc(text) + "</span>"; }
  var HEAD = "  " + s("mk", "▍") + s("b", "agentbook") + " " + s("f", "v1.0.0") + s("f", "  ·  ") + s("d", "Lumen") + "\n";
  function box(lines, color, title) {
    var w = 66, top;
    if (title) top = s(color, "╭─ ") + s(color + " b", title) + s(color, " " + "─".repeat(w - title.length - 3) + "╮");
    else top = s(color, "╭" + "─".repeat(w) + "╮");
    var out = "  " + top + "\n";
    lines.forEach(function (l) {
      var plain = l.replace(/<[^>]+>/g, "").replace(/&amp;/g, "&").replace(/&lt;/g, "<").replace(/&gt;/g, ">");
      out += "  " + s(color, "│") + "  " + l + " ".repeat(Math.max(0, w - 2 - [...plain].length)) + s(color, "│") + "\n";
    });
    return out + "  " + s(color, "╰" + "─".repeat(w) + "╯") + "\n";
  }
  function menu(sel) {
    var rows = [["◎", "Status", "Where the project stands and what's next"],
      ["⌕", "Find", "Search every decision, note and session log"],
      ["✓", "Check", "Is everything saved, pushed, logged and tested?"],
      ["❝", "Setup prompt", "Copy the prompt for any agent"],
      ["✚", "Doctor", "Look for problems on this computer"]];
    return rows.map(function (r, i) {
      var on = i === sel, title = (r[1] + "            ").slice(0, 14);
      return "  " + (on ? s("v", "▌ ") : "  ") + s("f", String(i + 1)) + "  " + (on ? s("v", r[0]) : s("d", r[0])) + "  " +
        (on ? s("v", title) : s("d", title)) + "  " + (on ? s("d", r[2]) : s("f", r[2]));
    }).join("\n") + "\n\n  " + s("v", "↑↓") + " " + s("f", "move") + "   " + s("v", "enter") + " " + s("f", "run") + "   " + s("v", "q") + " " + s("f", "quit") + "\n";
  }
  var card = box([
    s("hl", " private ") + " " + s("hl", " single layout ") + "  " + s("d", "updated today"),
    "",
    s("a", "NEXT STEP"),
    s("b", "Add a sign-up page to the website (see 2026-11-10 log)."),
    "",
    s("b", "14") + s("d", " decisions") + s("f", "  ·  ") + s("b", "9") + s("d", " sessions") + s("f", "  ·  ") + s("b", "61") + s("d", " questions"),
  ], "v", "Lumen");

  var demos = {
    home: [
      ["type", "agentbook"], ["wait", 300],
      ["out", "\n" + HEAD + "\n" + card + "\n"], ["menu", 0], ["wait", 900], ["menu", 1], ["wait", 700], ["menu", 2], ["wait", 700], ["menu", 1], ["wait", 1800],
    ],
    find: [
      ["type", "agentbook find phone"], ["wait", 300],
      ["out", "\n" + HEAD + "\n  " + s("b", "“phone”") + s("f", "  3 matches") + "\n\n"],
      ["out", "  " + s("v", "history/sessions/2026-11-02.md") + "\n"],
      ["out", s("f", "     12 ") + s("f", "│ ") + "### Q: Should we build a website or a " + s("a", "phone") + " app first?\n"],
      ["out", s("f", "     13 ") + s("f", "│ ") + "**A:** A website first. The " + s("a", "phone") + " app can wait.\n"],
      ["out", "  " + s("v", "DECISIONS.md") + "\n"],
      ["out", s("f", "     41 ") + s("f", "│ ") + "## 2026-11-02: Website first, " + s("a", "phone") + " app later\n\n"],
      ["wait", 2600],
    ],
    check: [
      ["type", "agentbook check"], ["wait", 300],
      ["out", "\n" + HEAD + "\n"],
      ["out", "  " + s("g", "✓") + " workspace: everything committed\n"],
      ["out", "  " + s("g", "✓") + " workspace: pushed\n"],
      ["out", "  " + s("g", "✓") + " STATUS.md updated today\n"],
      ["out", "  " + s("g", "✓") + " today's session log exists (" + s("v", "history/sessions/2026-11-11.md") + ")\n"],
      ["out", "  " + s("g", "✓") + " handbook is newer than the latest code commit\n"],
      ["spin", "Checking", 1400],
      ["out", "  " + s("g", "✓") + " tests pass\n\n"],
      ["out", box([s("g", "All clear"), s("d", "Everything is saved, pushed, logged and tested. Safe to stop."), s("g", "6 ✓") + s("f", "  ·  ") + s("f", "2.4s")], "g")],
      ["wait", 2600],
    ],
    init: [
      ["type", "agentbook init"], ["wait", 300],
      ["out", "\n  " + s("mk", "▍") + s("b", "agentbook") + " " + s("f", "v1.0.0") + s("f", "  ·  ") + s("d", "new workspace") + "\n\n"],
      ["out", s("v", "┃ ") + s("b", "Project name") + "\n" + s("v", "┃ ") + s("d", "Shown in the handbook and the agent rules.") + "\n" + s("v", "┃ ") + s("v", "> ") + "Lumen\n\n"],
      ["wait", 500],
      ["out", s("v", "┃ ") + s("b", "Which agents do you use?") + "\n" + s("v", "┃ ") + s("g", "◉ ") + "Claude Code  (CLAUDE.md)\n" + s("v", "┃ ") + s("g", "◉ ") + "Gemini CLI  (GEMINI.md)\n\n"],
      ["wait", 500],
      ["spin", "Creating the workspace", 1100],
      ["out", box([s("g", "Workspace ready") + "  " + s("f", "~/code/lumen"), "", s("a", "NEXT"),
        "1  Open the " + s("v", "lumen") + " folder in your coding agent",
        "2  Say: " + s("v", "“Read AGENTS.md and finish setting up.”"),
        "3  Then just work. Say " + s("v", "“wrap up”") + " when you're done."], "g")],
      ["wait", 2800],
    ],
  };

  var term = document.getElementById("term");
  var PROMPT = s("g", "➜") + " " + s("v", "lumen") + " ";
  var run = 0;

  function finalFrame(steps) {
    var html = PROMPT, typed = "";
    steps.forEach(function (st) {
      if (st[0] === "type") html += s("c", st[1]) + "\n";
      if (st[0] === "out") html += st[1];
      if (st[0] === "menu") typed = menu(st[1]);
    });
    return html + typed;
  }

  function play(name) {
    var id = ++run, steps = demos[name];
    if (reduced) { term.innerHTML = finalFrame(steps); return; }
    var base = PROMPT, menuHtml = "", i = 0;
    function render(extra) { term.innerHTML = base + menuHtml + (extra || ""); }
    function next() {
      if (id !== run) return;
      if (i >= steps.length) { setTimeout(function () { if (id === run) play(name); }, 400); return; }
      var st = steps[i++];
      if (st[0] === "type") {
        var n = 0;
        (function tick() {
          if (id !== run) return;
          render(s("c", st[1].slice(0, n)) + '<span class="cur"></span>');
          if (n++ < st[1].length) setTimeout(tick, 45 + Math.random() * 50);
          else { base += s("c", st[1]) + "\n"; render(); setTimeout(next, 250); }
        })();
      } else if (st[0] === "out") {
        base += st[1]; render(); setTimeout(next, 170);
      } else if (st[0] === "menu") {
        menuHtml = menu(st[1]); render(); next();
      } else if (st[0] === "spin") {
        var frames = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏", f = 0, start = Date.now();
        (function spin() {
          if (id !== run) return;
          var el = Date.now() - start;
          render("  " + s("v", frames[f++ % frames.length]) + " " + s("b", st[1]) + " " + s("f", (el / 1000).toFixed(1) + "s") + "\n");
          if (el < st[2]) setTimeout(spin, 80); else { render(); next(); }
        })();
      } else if (st[0] === "wait") {
        setTimeout(next, st[1]);
      }
    }
    term.innerHTML = "";
    next();
  }

  // Only animate while the terminal is on screen.
  var current = "home", started = false;
  tabs(document.querySelector(".term-tabs"), "data-demo", function (name) { current = name; if (started) play(name); });
  if ("IntersectionObserver" in window && !reduced) {
    new IntersectionObserver(function (es) {
      es.forEach(function (e) {
        if (e.isIntersecting && !started) { started = true; play(current); }
        else if (!e.isIntersecting && started) { started = false; run++; }
      });
    }, { threshold: 0.2 }).observe(term);
    term.innerHTML = finalFrame(demos.home);
  } else {
    started = true; play(current);
  }
})();
