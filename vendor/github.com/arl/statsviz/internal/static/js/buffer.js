// Buffer declares the Buffer class.
export default class Buffer {
    constructor(len, cap) {
        if (cap - len < 0) {
            throw "cap - len must be positive";
        }
        // TODO(arl): value using TypedArray rather than Array here
        this._buf = new Array(cap);
        this._pos = 0;
        this._len = len;
        this._cap = cap;
    }
    last() {
        if (this.length() == 0) {
            throw 'Cannot call last() on an empty Buffer';
        }
        return this._buf[this._pos];
    }
    push(pt) {
        if (this._pos >= this._cap) {
            // move data to the beginning of the buffer, effectively discarding
            // the cap-len oldest elements
            this._buf.copyWithin(0, this._cap - this._len);
            this._pos = this._len;
        }
        this._buf[this._pos] = pt;
        this._pos++;
    }
    length() {
        if (this._pos > this._len) {
            return this._len;
        }
        return this._pos;
    }

    // slice returns a slice of the len latest datapoints present in the buffer.
    slice(len) {
        // Cap the dimension of the returned slice to the data available
        if (len > this.length()) {
            len = this.length();
        }

        let start = this._pos - len;
        return this._buf.slice(start, start + len);
    }
};
