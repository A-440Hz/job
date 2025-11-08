import { useState } from "react";
export function About() {

    const [openPopup, setOpenPopup] = useState(false);

    return (
        <div className="mx-auto px-8 pt-4 max-w-8/10 justify-self-center justify-items-center mt-3">
            <h2 className='justify-self-center text-4xl'>About Page</h2>


        <button className="text-nowrap px-4 py-2 m-8 cursor-pointer border-3 bg-orange-300 backdrop-opacity-60 hover:opacity-75" type="button" onClick={() => setOpenPopup(true)}> view my resume </button>
        <p> {openPopup ? 'Popup is open' : 'Popup is closed'} </p>
        { openPopup && (
        <div id="myPopup" className="popup">
        <div className="popup-content no-select">
            <div className="popup-header">
            <div className="close" onClick={() => setOpenPopup(false)}>&times;</div>
            <div className="flex justify-between w-full">
                <span>
                <h4> select all images with </h4>
                <h2> motorcycles </h2>
                </span>
                <span>
                <img src="/moto.png"></img>
                </span>
            </div>
            </div>
            <div className="popup-frame">
            <img className="capcha-img" id="1" src="/1.jpg"></img>
            <img className="capcha-img" id="2" src="/2.jpg"></img>
            <img className="capcha-img" id="3" src="/3.jpg"></img>
            <img className="capcha-img" id="4" src="/4.jpg"></img>
            <img className="capcha-img" id="5" src="/5.jpg"></img>
            <img className="capcha-img" id="6" src="/6.jpg"></img>
            <img className="capcha-img" id="7" src="/7.jpg"></img>
            <img className="capcha-img" id="8" src="/8.jpg"></img>
            <img className="capcha-img" id="9" src="/9.jpg"></img>
            </div>
            <div className="popup-footer">
            <div className="err-msg"> Verification failed, please try again. </div>
            <div className="flex justify-between">
                <button className="btn" type="button" id="resetButton"> &#10227; </button>
                <button className="btn" type="button" id="submitButton"> verify </button>
            </div>
            </div>
        </div>
        <script src="/resume.js" type="module"></script>
        </div>
    )}
        </div>
    )
}