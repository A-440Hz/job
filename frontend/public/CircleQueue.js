export default class CQ {
    constructor() {
        this.size = 4;
        this.q = Array(this.size);
        this.tail = this.size-1;
        this.head = 0;
    }

    enqueue(v) {
        this.q[this.head] = v;
        this.head = (this.head + 1) % this.size;
        this.tail = (this.tail + 1) % this.size;
    }

    isValid() {
        let x = 1, y = 0, z = "";
        for (let i = 0; i < this.size; i++) {
            let a = this.q[(this.tail + i+1) % this.size];
            y += Number(a);
            x *= a;
            z += a;
            if (i % 2 != 0 && a * i != x) {
                return false;
            }
        }
        return y - x == z[x%2];
    }
}