import React, { createContext, useContext, useEffect, useState } from 'react';
import { fetchTrackerData } from './api/tracker';

type TrackerDataContextType = {
    data: any;
    error: string | null;
    refreshData: () => void;
};

const TrackerDataContext = createContext<TrackerDataContextType>({
    data: null,
    error: null,
    refreshData: () => {},
});

export function useTrackerData() {
    return useContext(TrackerDataContext);
}

export function TrackerDataProvider({ children }: { children: React.ReactNode }) {
    const [data, setData] = useState<any>(null);
    const [error, setError] = useState<string | null>(null);

    const fetchData = () => {
        fetchTrackerData()
            .then(setData)
            .catch((err) => setError(err.message));
    };

    useEffect(fetchData, []);

    return (
        <TrackerDataContext.Provider value={{ data, error, refreshData: fetchData }}>
            {children}
        </TrackerDataContext.Provider>
    );
}
