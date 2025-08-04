import React, { createContext, useContext, useEffect, useState } from 'react';

// Create a context
const ScreenSizeContext = createContext(false);

// Custom hook to use the ScreenSizeContext
export const useScreenSize = () => {
    return useContext(ScreenSizeContext);
};

export const ScreenSizeProvider = ({ children }:{children:React.ReactNode}) => {
    const [isDesktop, setIsDesktop] = useState(false);

    useEffect(() => {
        const match = window.matchMedia('(min-width: 768px)');
        setIsDesktop(match.matches);

        const handleChange = (e:any) => {
            setIsDesktop(e.matches);
        };

        match.addEventListener('change', handleChange);

        // Cleanup listener on unmount
        return () => {
            match.removeEventListener('change', handleChange);
        };
    }, []);

    return (
        <ScreenSizeContext.Provider value={isDesktop}>
            {children}
        </ScreenSizeContext.Provider>
    );
};
