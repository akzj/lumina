// Lumina Web Client — xterm.js + WebSocket bridge
(function() {
    'use strict';

    var statusEl = document.getElementById('status');
    function setStatus(msg) { if (statusEl) statusEl.textContent = msg; }

    // Create terminal
    var term = new Terminal({
        cursorBlink: true,
        cursorStyle: 'block',
        theme: {
            background: '#1e1e2e',
            foreground: '#cdd6f4',
            cursor: '#f5e0dc',
            selectionBackground: '#585b70',
            black: '#45475a',
            red: '#f38ba8',
            green: '#a6e3a1',
            yellow: '#f9e2af',
            blue: '#89b4fa',
            magenta: '#f5c2e7',
            cyan: '#94e2d5',
            white: '#bac2de',
            brightBlack: '#585b70',
            brightRed: '#f38ba8',
            brightGreen: '#a6e3a1',
            brightYellow: '#f9e2af',
            brightBlue: '#89b4fa',
            brightMagenta: '#f5c2e7',
            brightCyan: '#94e2d5',
            brightWhite: '#a6adc8'
        },
        fontFamily: 'JetBrains Mono, Cascadia Code, Fira Code, Menlo, monospace',
        fontSize: 14,
        lineHeight: 1.1,
        allowProposedApi: true
    });

    var fitAddon = new FitAddon.FitAddon();
    term.loadAddon(fitAddon);
    term.loadAddon(new WebLinksAddon.WebLinksAddon());
    term.open(document.getElementById('terminal'));
    fitAddon.fit();

    // WebSocket connection
    var proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    var wsUrl = proto + '//' + location.host + '/ws';
    var ws = null;
    var reconnectDelay = 1000;

    function connect() {
        ws = new WebSocket(wsUrl);
        ws.binaryType = 'arraybuffer';

        ws.onopen = function() {
            setStatus('connected');
            reconnectDelay = 1000;
            // Send initial terminal size
            ws.send(JSON.stringify({
                type: 'resize',
                cols: term.cols,
                rows: term.rows
            }));
        };

        ws.onmessage = function(event) {
            if (event.data instanceof ArrayBuffer) {
                term.write(new Uint8Array(event.data));
            } else {
                term.write(event.data);
            }
        };

        ws.onclose = function() {
            setStatus('disconnected — reconnecting...');
            setTimeout(connect, reconnectDelay);
            reconnectDelay = Math.min(reconnectDelay * 2, 10000);
        };

        ws.onerror = function() {
            setStatus('error');
        };
    }

    // Send keystrokes to server
    term.onData(function(data) {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(data);
        }
    });

    // Send binary data (for special keys)
    term.onBinary(function(data) {
        if (ws && ws.readyState === WebSocket.OPEN) {
            var buf = new Uint8Array(data.length);
            for (var i = 0; i < data.length; i++) {
                buf[i] = data.charCodeAt(i) & 0xff;
            }
            ws.send(buf.buffer);
        }
    });

    // Handle resize
    var resizeTimer = null;
    window.addEventListener('resize', function() {
        clearTimeout(resizeTimer);
        resizeTimer = setTimeout(function() {
            fitAddon.fit();
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({
                    type: 'resize',
                    cols: term.cols,
                    rows: term.rows
                }));
            }
        }, 100);
    });

    // Start connection
    connect();
})();
