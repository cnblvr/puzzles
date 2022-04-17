'use strict';

class Sudoku {
    #_object;
    #_keyboard;
    #isWin = false;
    #gameID;
    #ws;
    #cndMode = false;

    #_option_useHighlights = undefined;
    #_option_useCandidates = undefined;

    constructor(param) {
        if (!param)
            throw 'sudoku: parameters not defined';
        if (!param.selector || typeof param.selector !== 'string')
            throw 'sudoku: required parameter \'selector\' is not defined or not string';
        if (!param.gameID || typeof param.gameID !== 'string')
            throw 'sudoku: required parameter \'gameID\' is not defined or not string';
        this.#gameID = param.gameID;
        this.#_object = document.querySelector(param.selector);
        if (!this.#_object)
            throw 'sudoku: object by parameter \'selector\' not found';
        if (param.allowEditing) {
            if (typeof param.allowEditing !== 'boolean')
                throw 'sudoku: parameter \'allowEditing\' is not boolean';
            if (param.keyboardSelector) {
                if (typeof param.keyboardSelector !== 'string')
                    throw 'sudoku: parameter \'keyboardSelector\' is not string';
                this.#_keyboard = document.querySelector(param.keyboardSelector);
                if (!this.#_keyboard)
                    throw 'sudoku: object by parameter \'keyboardSelector\' not found';
            }
        }
        if (param.options) {
            if (param.options.useHighlights) {
                this.#_option_useHighlights = document.querySelector(param.options.useHighlights);
                if (!this.#_option_useHighlights)
                    throw 'sudoku: object by parameter \'options.useHighlights\' not found';
                this.#_option_useHighlights.addEventListener('change', (e) => {
                    this.#ws.send('setUserPreferences', { use_highlights: e.currentTarget.checked });
                    if (e.currentTarget.checked) {
                        let _active = this.#_object.querySelector('.sud-cll.active');
                        if (!_active) return;
                        this.#_setHighlights(_active.querySelector('.sud-dgt').textContent, _active);
                    } else {
                        this.#_resetHighlights();
                    }
                });
            }
            if (param.options.useCandidates) {
                this.#_option_useCandidates = document.querySelector(param.options.useCandidates);
                if (!this.#_option_useCandidates)
                    throw 'sudoku: object by parameter \'options.useCandidates\' not found';
                if (this.#_isUseCandidates()) this.#_object.classList.remove('hide-candidates');
                else this.#_object.classList.add('hide-candidates');
                this.#_option_useCandidates.addEventListener('change', (e) => {
                    this.#ws.send('setUserPreferences', { use_candidates: e.currentTarget.checked });
                    if (e.currentTarget.checked) {
                        this.#_object.classList.remove('hide-candidates');
                    } else {
                        this.#_object.classList.add('hide-candidates');
                    }
                });
            }
        }

        this.#_object.classList.add('sudoku');
        for (let row = 0; row < 9; row++) {
            let _row = document.createElement('div');
            _row.classList.add('sud-row');
            for (let col = 0; col < 9; col++) {
                let _cell = document.createElement('div');
                _cell.classList.add('sud-cll', 'is-cnd');
                // create digit field
                let _dgt = document.createElement('div');
                _dgt.classList.add('sud-dgt');
                _cell.appendChild(_dgt);
                // create table of candidates
                let _cnd = document.createElement('div');
                _cnd.classList.add('sud-cnd');
                for (let idx = 1; idx <= 9; idx++) {
                    let _cndItem = document.createElement('div');
                    _cndItem.classList.add('hidden');
                    _cndItem.textContent = ''+idx;
                    _cnd.appendChild(_cndItem);
                }
                _cell.appendChild(_cnd);
                _row.appendChild(_cell);
            }
            this.#_object.appendChild(_row);
        }

        if (this.#_keyboard) {
            this.#_keyboard.classList.add('keyboard');
            let createBtn = (label, event, id) => {
                let btn = document.createElement('div');
                btn.classList.add('kb-btn');
                btn.textContent = label;
                if (id) btn.id = id;
                btn.addEventListener('click', event);
                this.#_keyboard.appendChild(btn);
                return btn;
            }
            createBtn( 'c', (e) => {
                this.#toggleCandidateMode();
            }, 'cnd-mode').title = 'press [C] to switch mode; press [Shift]+[digit] to set the candidate';
            createBtn( '⨯', (e) => {
                this.#cndMode?
                    this.#toggleCandidateInActive('0'):
                    this.#placeDigitInActive('0');
            }).title = 'press [Backspace], [Space] or [0] to remove the digit; press [Shift]+[one of the previous keys] to remove all candidates';
            for (let digit = 1; digit <= 9; digit++) {
                createBtn(digit, (e) => {
                    this.#cndMode?
                        this.#toggleCandidateInActive(''+digit):
                        this.#placeDigitInActive(''+digit);
                });
            }
        }

        if (param.allowEditing) {
            this.#_object.querySelectorAll('.sud-cll').forEach((_cell) => {
                _cell.addEventListener('mouseup', (e) => {
                    this.#setActive(_cell);
                });
            });

            document.addEventListener('keydown', (e) => {
                if (e.defaultPrevented) {
                    return;
                }
                switch (e.code) {
                    case 'ShiftLeft':
                    case 'ShiftRight':
                        this.#toggleCandidateMode(true);
                        break;
                }
            });
            document.addEventListener('keyup', (e) => {
                if (e.defaultPrevented) {
                    return;
                }
                let _a = this.#_object.querySelector('.sud-cll.active');
                switch (e.code) {
                    case 'ArrowUp':    this.#setActive(_a, 'up');    break;
                    case 'ArrowRight': this.#setActive(_a, 'right'); break;
                    case 'ArrowDown':  this.#setActive(_a, 'down');  break;
                    case 'ArrowLeft':  this.#setActive(_a, 'left');  break;
                    case 'Digit0':
                    case 'Numpad0':
                    case 'Space':
                    case 'Backspace':
                        this.#cndMode?
                            this.#toggleCandidateInActive('0'):
                            this.#placeDigitInActive('0');
                        break;
                    case 'KeyC':
                        this.#toggleCandidateMode();
                        break;
                    case 'ShiftLeft':
                    case 'ShiftRight':
                        this.#toggleCandidateMode(false);
                        break;
                }
                let digit = undefined;
                switch (true) {
                case 'Digit1' <= e.code && e.code <= 'Digit9':
                    digit = e.code.replace('Digit', '');
                    break
                case 'Numpad1' <= e.code && e.code <= 'Numpad9':
                    digit = e.code.replace('Numpad', '');
                    break;
                }
                if (digit) this.#cndMode?
                    this.#toggleCandidateInActive(digit):
                    this.#placeDigitInActive(e.key);
            });
        }

        this.#_object.addEventListener('apiReady', () => {
            this.#ws.send('getPuzzle', {
                game_id: this.#gameID,
            });
        }, {once: true});

        this.#_object.addEventListener('api_getPuzzle', (e) => {
            let body = e.detail.body;
            let puzzle = body.is_new ? body.puzzle : body.state_puzzle;
            let candidates = body.is_new ? body.candidates : body.state_candidates;
            this.#_object.querySelectorAll('.sud-row').forEach((_row, row) => {
                _row.querySelectorAll('.sud-cll').forEach((_cell, col) => {
                    this.#placeDigit(_cell, '0', true);
                    let d = puzzle[row * 9 + col];
                    if ('1' <= d && d <= '9') {
                        this.#placeDigit(_cell, d, true);
                        if (body.puzzle[row * 9 + col] !== '.')
                            _cell.classList.add('hint');
                    }
                    if (candidates.base) {
                        this.#setCandidatesFor(_cell, candidates.base[this.#stringifyPoint(row, col)]);
                    }
                });
            });
            if (!body.is_new) {
                this.#deleteErrors();
                this.#setErrors(body.errors, body.errorsCandidates);
            }
            if (body.is_win) this.#isWin = true;
        });

        this.#_object.addEventListener('api_makeStep', (e) => {
            let body = e.detail.body;
            this.#deleteErrors();
            if (body.win) {
                this.#isWin = true;
                alert('win'); // TODO
                return;
            }
            this.#setErrors(body.errors, body.errorsCandidates);
        });
    }

    dispatchEvent(ce) {
        this.#_object.dispatchEvent(ce);
    }

    connectWS(ws) {
        this.#ws = ws;
    }

    #placeDigit(_cell, digit, notMakeStep) {
        if (this.#isWin) return;
        if (!_cell || _cell.classList.contains('hint')) return;
        let _digit = _cell.querySelector('.sud-dgt');
        let oldDigit = _digit.textContent===''?'0':_digit.textContent;
        _cell.classList.remove('is-dgt', 'is-cnd');
        if (digit === '0') {
            _digit.textContent = '';
            _cell.classList.add('is-cnd');
        } else {
            _digit.textContent = digit;
            _cell.classList.add('is-dgt');
        }
        if (!notMakeStep) this.#_setHighlights(_cell.querySelector('.sud-dgt').textContent, _cell);
        let point = this.#stringifyPoint(this.#getIndex(_cell.parentElement), this.#getIndex(_cell));
        if (!notMakeStep && oldDigit !== digit) this.#apiMakeStep(digit === '0'?'del_digit':'set_digit', point, digit);
    }

    #placeDigitInActive(digit, notMakeStep) {
        if (this.#isWin) return;
        this.#placeDigit(this.#_object.querySelector('.sud-cll.active'), digit, notMakeStep);
    }

    #_isUseCandidates() {
        return !!(this.#_option_useCandidates && this.#_option_useCandidates.checked);
    }

    #setCandidatesFor(_cell, cands) {
        if (!_cell || !cands) return;
        _cell.querySelectorAll('.sud-cnd div').forEach((_div) => {
            if (cands.includes(_div.textContent.charCodeAt(0)-'0'.charCodeAt(0))) {
                _div.classList.remove('hidden');
            } else {
                _div.classList.add('hidden');
            }
        });
    }

    #setActive(_cell, dir) {
        if (!_cell) {
            _cell = this.#_object.querySelectorAll('.sud-row').item(9/2).querySelectorAll('.sud-cll').item(9/2);
            dir = undefined;
            if (!_cell) return;
        }
        let _row = _cell.closest('.sud-row');
        switch (dir) {
            case 'up':
                let _prev = _row.previousElementSibling;
                if (!_prev) return;
                _cell = _prev.querySelectorAll('.sud-cll').item(this.#getIndex(_cell));
                break;
            case 'right':
                _cell = _cell.nextElementSibling; break;
            case 'down':
                let _next = _row.nextElementSibling;
                if (!_next) return;
                _cell = _next.querySelectorAll('.sud-cll').item(this.#getIndex(_cell));
                break;
            case 'left':
                _cell = _cell.previousElementSibling; break;
        }
        if (!_cell) return;
        let isAlready = _cell.classList.contains('active');
        this.#_object.querySelectorAll('.sud-cll.active').forEach((_active) => {
            _active.classList.remove('active');
        });
        if (!isAlready) {
            _cell.classList.add('active');
            this.#_setHighlights(_cell.querySelector('.sud-dgt').textContent, _cell);
        } else this.#_resetHighlights();
    }

    #_isUseHighlights() {
        return !!(this.#_option_useHighlights && this.#_option_useHighlights.checked);
    }

    #_resetHighlights() {
        this.#_object.querySelectorAll('.sud-cll').forEach((_cell) => {
            _cell.classList.remove('hl');
            _cell.querySelectorAll('.sud-cnd div').forEach((_cnd) => {
                _cnd.classList.remove('hl');
            });
        });
    }

    #_setHighlights(digit, exclude) {
        this.#_resetHighlights();
        if (!this.#_isUseHighlights()) return;
        if (!digit || digit === '') return;
        this.#_object.querySelectorAll('.sud-cll').forEach((_cell) => {
            if (exclude === _cell) return;
            let dhl = _cell.querySelector('.sud-dgt').textContent;
            if (dhl === '') {
                _cell.querySelectorAll('.sud-cnd div').forEach((_cnd) => {
                    if (_cnd.textContent === digit) _cnd.classList.add('hl');
                });
            } else if (dhl === digit) {
                _cell.classList.add('hl');
            }
        });
    }

    #toggleCandidateMode(state) {
        if (state && typeof state === 'boolean') {
            this.#cndMode = state;
        } else {
            this.#cndMode = !this.#cndMode;
        }
        let cndMode = document.querySelector('#cnd-mode');
        if (this.#cndMode) cndMode.classList.add('active');
        else cndMode.classList.remove('active');
    }

    #toggleCandidate(_cell, digit) {
        if (this.#isWin) return;
        if (!_cell || _cell.classList.contains('hint') || !_cell.classList.contains('is-cnd')) return;
        let point = this.#stringifyPoint(this.#getIndex(_cell.parentElement), this.#getIndex(_cell));
        _cell.querySelectorAll('.sud-cnd div').forEach((_cnd) => {
            if (digit === '0') { _cnd.classList.add('hidden'); return; }
            if (_cnd.textContent !== digit) return;
            if (_cnd.classList.contains('hidden')) {
                _cnd.classList.remove('hidden');
                this.#apiMakeStep('set_cand', point, digit);
            } else {
                _cnd.classList.add('hidden');
                this.#apiMakeStep('del_cand', point, digit);
            }
            this.#_setHighlights(digit, undefined);
        });
    }

    #toggleCandidateInActive(digit) {
        if (this.#isWin) return;
        this.#toggleCandidate(this.#_object.querySelector('.sud-cll.active'), digit);
    }

    #deleteErrors() {
        this.#_object.querySelectorAll('.sud-cll').forEach((_cell) => {
            _cell.classList.remove('error');
        });
        this.#_object.querySelectorAll('.sud-cnd div').forEach((_cnd) => {
            _cnd.classList.remove('error');
        });
    }

    #setErrors(errors, errorsCandidates) {
        errors = this.#parsePoints(errors);
        this.#_object.querySelectorAll('.sud-row').forEach((_row, row) => {
            _row.querySelectorAll('.sud-cll').forEach((_cell, col) => {
                errors.forEach((p) => {
                    if (p.row === row && p.col === col) {
                        _cell.classList.add('error');
                    }
                });
                if (errorsCandidates.base) {
                    let candidates = errorsCandidates.base[this.#stringifyPoint(row, col)];
                    if (candidates) {
                        candidates.forEach((cand) => {
                            _cell.querySelectorAll('.sud-cnd div')[cand - 1].classList.add('error');
                        });
                    }
                }
            });
        });
    }

    #apiMakeStep(type, point, digit) {
        this.#ws.send('makeStep', {
            game_id: this.#gameID,
            step: {
                type: type,
                point: point,
                digit: digit.charCodeAt(0)-'0'.charCodeAt(0),
            },
        });
    }

    #getIndex(_node) {
        let index = 0;
        while (_node = _node.previousElementSibling) {
            index++;
        }
        return index;
    }

    #parsePoints(points) {
        let out = [];
        if (!points) return out;
        points.forEach((p) => {
            out = out.concat([{
                row: p[0].charCodeAt(0)-'a'.charCodeAt(0),
                col: parseInt(p[1])-1,
            }]);
        });
        return out;
    }

    #stringifyPoint(row, col) {
        return String.fromCharCode((row)+'a'.charCodeAt(0)) + (col+1);
    }
}
