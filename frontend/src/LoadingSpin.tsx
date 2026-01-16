/**
 * @component
 * @returns a React component for displaying a loading spinner
 */
function LoadingSpin() {  
    return (
        <img 
            className="w-6 mr-1 rounded-xl motion-safe:animate-[spin_.6s_linear_infinite]"
            src="/squid-panic.png" 
            alt="Loading..."
        />
    );
};

export default LoadingSpin;