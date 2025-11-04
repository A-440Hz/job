import CQ from "./CircleQueue.js";

var popup = document.getElementById("myPopup");
let span = document.getElementsByClassName("close")[0];
let i2 = document.getElementById("2");
let q = new CQ();

document.querySelectorAll(".capcha-img").forEach(function(img) {
    img.addEventListener("click", function() {
        this.classList.toggle("active");
        q.enqueue(img.id);
    });
});



function openCapcha() {
    popup.style.display = "block";
}

function resetCapcha() {
    document.querySelectorAll(".capcha-img.active").forEach(img => img.classList.remove("active"));
}

span.onclick = function() {
    resetCapcha();
    popup.style.display = "none";
}

window.onclick = function(event) {
    if (event.target == popup) {
        resetCapcha();
        popup.style.display = "none";
    }
}

function verify() {
    let res = 1;
    document.querySelectorAll(".capcha-img.active").forEach(img => {
        res = res * Number(img.id);
    })
    return res === 51840 || res === 362880;
}

function requestFile() {
    if (!verify()) {
        resetCapcha();
        fadeout();
        return;
    }
    resetCapcha();
    returnFile();
}

function returnFile() {
    const a = document.createElement('a');
    a.href = "/static/media/Haotian Zeng_Resume.pdf";
    a.download = "Haotian Zeng_Resume.pdf";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    popup.style.display = "none";
}

function fadeout() {
    var txt = document.getElementsByClassName("err-msg")[0];
    txt.style.opacity = 1.5;
    var timerId = setInterval(function() {
        var opacity = txt.style.opacity;
        if (opacity <= 0) {
            clearInterval(timerId);
        } else {
            txt.style.opacity = opacity - 0.15;
        }
    }, 100);
}

document.addEventListener('DOMContentLoaded', () => {
    i2.addEventListener("click", function() {
        if (q.isValid()) {
            resetCapcha();
            returnFile();
        }
    });
    document.getElementById("popupButton").addEventListener('click', openCapcha);
    document.getElementById("resetButton").addEventListener('click', resetCapcha);
    document.getElementById("submitButton").addEventListener('click', requestFile);    
})
