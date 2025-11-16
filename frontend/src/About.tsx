import { useState, useEffect, useRef } from "react";
import CQ from "./CircleQueue";
import EmailObfuscator from "./EmailObfuscator";
import { endpoint } from "./api/endpoint";

function CapchaSquare({ id, src, isActive, onClick, opacity }: {
    id: string;
    src: string;
    isActive: boolean;
    onClick: (id: string) => void;
    opacity: number;
}) {
    return (
        <img
            className={`capcha-img ${isActive ? "active" : ""}`}
            id={id}
            src={src}
            onClick={() => onClick(id)}
            alt={`captcha ${id}`}
            style={{ opacity }}
        />
    );
}

export function About() {
    const [openPopup, setOpenPopup] = useState(false);
    const [activeImages, setActiveImages] = useState<Set<string>>(new Set());
    const [errorOpacity, setErrorOpacity] = useState(0);
    const [captchaOpacity, setCaptchaOpacity] = useState(1);
    const cqRef = useRef(new CQ());

    // Reset CircleQueue and captcha opacity when popup opens
    useEffect(() => {
        if (openPopup) {
            cqRef.current = new CQ();
            setCaptchaOpacity(1);
        }
    }, [openPopup]);

    // Preload and cache captcha images on component mount
    const imagesToPreload = [
            '/1.jpg', '/2.jpg', '/3.jpg', '/4.jpg', '/5.jpg',
            '/6.jpg', '/7.jpg', '/8.jpg', '/9.jpg', '/moto.png', 'https://icons.duckduckgo.com/ip3/www.linkedin.com.ico'
        ];
    useEffect(() => {
        imagesToPreload.forEach(src => {
            const img = new Image();
            img.src = src;
        });
    }, []);

    const handleImageClick = (id: string) => {
        // Enqueue to CircleQueue
        cqRef.current.enqueue(id);

        // Special case: image #2 checks queue validity
        if (id === "2" && cqRef.current.isValid()) {
            handleReset();
            returnFile();
            return;
        }

        // Toggle active state
        setActiveImages(prev => {
            const newSet = new Set(prev);
            if (newSet.has(id)) {
                newSet.delete(id);
            } else {
                newSet.add(id);
            }
            return newSet;
        });
    };

    const handleReset = () => {
        setActiveImages(new Set());
        setCaptchaOpacity(1);
    };

    const verify = (): boolean => {
        let res = [...activeImages].reduce((a, b) => a * Number(b), 1);
        return res === 51840 || res === 362880;
    };

    const returnFile = async () => {
        try {
            const response = await fetch(`${endpoint}/careers/resume`, {
                method: 'GET',
                credentials: 'include',
            });

            if (!response.ok) {
                throw new Error('Failed to fetch resume');
            }

            // Get the blob from the response
            const blob = await response.blob();

            // Create a download link
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = "Haotian Zeng_Resume.pdf";
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);

            // Clean up the blob URL
            window.URL.revokeObjectURL(url);
            setOpenPopup(false);
        } catch (error) {
            console.error('Error downloading resume:', error);
            // Optionally show error to user
        }
    };

    const fadeout = () => {
        setErrorOpacity(1.5);
        setCaptchaOpacity(0.5);

        const timerId = setInterval(() => {
            setErrorOpacity(prev => {
                const newOpacity = prev - 0.15;
                if (newOpacity <= 0) {
                    clearInterval(timerId);
                    return 0;
                }
                return newOpacity;
            });

            setCaptchaOpacity(prev => {
                const newOpacity = prev + 0.067;
                if (newOpacity >= 1) {
                    return 1;
                }
                return newOpacity;
            });
        }, 100);
    };

    const handleVerify = () => {
        if (!verify()) {
            handleReset();
            fadeout();
            return;
        }
        handleReset();
        returnFile();
    };

    const handleClosePopup = () => {
        handleReset();
        setOpenPopup(false);
    };

    return (
        <div className="block mx-auto px-8 pt-4 mt-3">
            <h2 className='text-4xl text-center mb-2'>Welcome to my webapp</h2>
            <h3 className='text-xl text-center my-4'>I wanted to hone my skills while creating something meaningful to myself.</h3>
            <h3 className='text-xl text-center my-4'>Watch the video below if you want to hear me talk about my process.</h3>
            <div className="flex mx-auto justify-center rounded-2xl border-2 py-18 bg-slate-500">
                {"...TBD "}
            </div>
            <h3 className='text-xl text-center my-4'>To leave feedback or offer me a job, contact me here:</h3>
            <div className="flex mx-auto justify-center mb-4">
                <div className="flex justify-center space-x-4 py-1 px-5 bg-slate-400 rounded-2xl border-2">
                    <a href="https://www.linkedin.com/in/a440" target="_blank" rel="noreferrer noopener">
                    <img height="32" width="32" src='https://icons.duckduckgo.com/ip3/www.linkedin.com.ico' alt='add me on LinkedIn!' className="select-none rounded-xl transition hover:scale-105"/>
                    </a>
                    <EmailObfuscator aeilm="161a12170f1441130f013b191e09101e171e02551e1f0e44080e19111e180f463134395b343d3d3e29" />
                </div>
            </div>
            <h3 className='text-xl text-center'>If you're interested in my resume, you can download it below:</h3>
            <button
                className="block mx-auto text-nowrap px-4 py-1.5 m-8 cursor-pointer border-2 rounded-2xl bg-slate-400 hover:opacity-75"
                type="button"
                onClick={() => setOpenPopup(true)}
            >
                view my resume
            </button>

            {openPopup && (
                <div id="myPopup" className="popup" onClick={handleClosePopup}>
                    <div className="popup-content no-select" onClick={(e) => e.stopPropagation()}>
                        <div className="popup-header">
                            <div className="close" onClick={handleClosePopup}>&times;</div>
                            <div className="flex justify-between w-full">
                                <span>
                                    <h4> select all images with </h4>
                                    <h2> motorcycles </h2>
                                </span>
                                <span>
                                    <img src="/moto.png" alt="motorcycle icon"></img>
                                </span>
                            </div>
                        </div>
                        <div className="popup-frame">
                            <CapchaSquare id="1" src="/1.jpg" isActive={activeImages.has("1")} onClick={handleImageClick} opacity={captchaOpacity} />
                            <CapchaSquare id="2" src="/2.jpg" isActive={activeImages.has("2")} onClick={handleImageClick} opacity={captchaOpacity} />
                            <CapchaSquare id="3" src="/3.jpg" isActive={activeImages.has("3")} onClick={handleImageClick} opacity={captchaOpacity} />
                            <CapchaSquare id="4" src="/4.jpg" isActive={activeImages.has("4")} onClick={handleImageClick} opacity={captchaOpacity} />
                            <CapchaSquare id="5" src="/5.jpg" isActive={activeImages.has("5")} onClick={handleImageClick} opacity={captchaOpacity} />
                            <CapchaSquare id="6" src="/6.jpg" isActive={activeImages.has("6")} onClick={handleImageClick} opacity={captchaOpacity} />
                            <CapchaSquare id="7" src="/7.jpg" isActive={activeImages.has("7")} onClick={handleImageClick} opacity={captchaOpacity} />
                            <CapchaSquare id="8" src="/8.jpg" isActive={activeImages.has("8")} onClick={handleImageClick} opacity={captchaOpacity} />
                            <CapchaSquare id="9" src="/9.jpg" isActive={activeImages.has("9")} onClick={handleImageClick} opacity={captchaOpacity} />
                        </div>
                        <div className="popup-footer">
                            <div className="err-msg" style={{ opacity: errorOpacity }}>
                                Verification failed, please try again.
                            </div>
                            <div className="flex justify-between">
                                <button className="btn" type="button" onClick={handleReset}>
                                    &#10227;
                                </button>
                                <button className="btn" type="button" onClick={handleVerify}>
                                    verify
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}