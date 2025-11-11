// inspiration from https://spencermortensen.com/articles/email-obfuscation/
// if i wanted more security i'd use RSA or capcha or something.
// i want to see if i can circumvent crawlers and spam
// this is just for fun i have my personal info on a public pdf anyway

import { useEffect, useState } from "react";

class Xor {
    private key: number;

    constructor(key: number) {
        this.key = key;
    }

    encode(input: string): string {
        let ret = '';
        for (let i = 0; i < input.length; i++) {
            const textIn = input.charCodeAt(i);
            const textOut = textIn ^ this.key;
            ret += this.toHex(textOut);
        }
        return ret;
    }

    decode(input: string): string {
        let ret = '';
        for (let i = 0; i < input.length; i += 2) {
            const hexIn = this.fromHex(input, i);
            const hexOut = hexIn ^ this.key;
            ret += String.fromCharCode(hexOut);
        }
        return ret;
    }

    toHex(input: number): string {
        return input.toString(16).padStart(2, '0');
    }

    fromHex(hex: string, i: number): number {
        const sequence = hex.substr(i, 2);
        return parseInt(sequence, 16);
    }
}

const EmailObfuscator = ({ aeilm }: { aeilm: string }) => {
    const xor = new Xor(123);
    const [isObfuscated, setIsObfuscated] = useState(true);
    const [displayedEmail, setDisplayedEmail] = useState(aeilm);

    const setEmailObfuscation = (obfuscate: boolean) => {
        if (obfuscate) {
            setDisplayedEmail(aeilm);
            if (isObfuscated) return;
            setIsObfuscated(true);
        } else {
            setDisplayedEmail(xor.decode(aeilm));
            if (!isObfuscated) return;
            setIsObfuscated(false);
        }
    }
    
    return (
        <a
            href={displayedEmail}
            onMouseOver={() => setEmailObfuscation(false)}
            onMouseOut={() => setEmailObfuscation(true)}
            onFocus={() => setEmailObfuscation(false)}
            onBlur={() => setEmailObfuscation(true)}
            className={`flex select-none mr-4 ${isObfuscated && "cursor-not-allowed"}`}
            target="_blank"
            rel="noreferrer noopener"
        >
            <img height="32" width="32" src='/envelope.png' alt="email icon" />
        </a>
    );
};

export default EmailObfuscator;