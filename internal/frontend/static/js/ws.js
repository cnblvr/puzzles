'use strict';

class WS {
    #ws;
    #debug = false;
    #sudoku;

    constructor(param) {
        if (!param)
            throw 'ws: parameters not defined';
        if (!param.url || typeof param.url !== 'string')
            throw 'ws: required parameter \'url\' is not defined or not string';
        if (param.debug) {
            if (typeof param.debug !== 'boolean')
                throw 'ws: parameter \'debug\' is not boolean';
            if (param.debug === true)
                this.#debug = true
        }
        if (!param.sudoku)
            throw 'ws: required parameter \'sudoku\' is not defined';
        this.#sudoku = param.sudoku;
        this.#connect(param.url);
    }

    send(method, body) {
        if (!body || !method) return;
        let msg = JSON.stringify({
            method: method,
            body: body,
        });
        if (this.#debug) console.log('ws: send message:', msg);
        this.#ws.send(msg);
    }

    #connect(url) {
        this.#ws = new WebSocket(url);
        this.#ws.onopen = (e) => {
            if (this.#debug) console.log('ws: open connection');
            this.#sudoku.dispatchEvent(new CustomEvent('apiReady'));
        }
        this.#ws.onclose = (e) => {
            if (this.#debug) console.log('ws: close connection');
            this.#ws = undefined;
            // reconnect
            setTimeout(()=>{this.#connect(url)}, 1000);
        }
        this.#ws.onmessage = (e) => {
            if (this.#debug) console.log('ws: receive message:', e.data);
            let msg = JSON.parse(e.data);
            if (!msg.method) return;
            if (msg.error) {
                console.error('api', msg.method, 'error:', msg.error);
                return;
            }
            this.#sudoku.dispatchEvent(new CustomEvent('api_'+msg.method, {detail: msg}));
        }
        this.#ws.onerror = (e) => {
            console.error('ws: error '+e.code+':', e.reason, e);
            this.#ws.close();
        }
    }
}