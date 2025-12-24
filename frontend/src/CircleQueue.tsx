export default class CQ {
    private size: number;
    private q: (string | undefined)[];
    private tail: number;
    private head: number;

    constructor() {
        this.size = 4;
        this.q = Array(this.size);
        this.tail = this.size - 1;
        this.head = 0;
    }

    enqueue(v: string) {
        this.q[this.head] = v;
        this.head = (this.head + 1) % this.size;
        this.tail = (this.tail + 1) % this.size;
    }

    isValid(): boolean {
        let x = 1, y = 0, z = "";
        for (let i = 0; i < this.size; i++) {
            let a = this.q[(this.tail + i + 1) % this.size];
            y += Number(a);
            x *= Number(a);
            z += a;
            if (i % 2 != 0 && Number(a) * i != x) {
                return false;
            }
        }
        return y - x == Number(z[x % 2]);
    }
}
