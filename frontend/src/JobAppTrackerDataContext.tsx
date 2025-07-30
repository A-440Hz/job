import React, { createContext, useContext, useEffect, useState } from 'react';
import { fetchTrackerData } from './api/tracker';

type TrackerDataContextType = {
    user: any;
    tracker: any;
    error: string | null;
    refreshData: () => void;
    setTracker: (d: any) => void;
};

const TrackerDataContext = createContext<TrackerDataContextType>({
    user: null,
    tracker: null,
    error: null,
    refreshData: () => {},
    setTracker: () => {},
});

export function useTrackerData() {
    return useContext(TrackerDataContext);
}

export function TrackerDataProvider({ children }: { children: React.ReactNode }) {
    const [user, setUser] = useState<any>(null);
    const [tracker, setTracker] = useState<any>(null);
    const [error, setError] = useState<string | null>(null);

    const fetchData = () => {
        fetchTrackerData()
            .then((data) => {
                setUser(data.user);
                setTracker(data.tracker);
            })
            .catch((err) => setError(err.message));
    };

    useEffect(fetchData, []);

    return (
        <TrackerDataContext.Provider value={{ user, tracker, error, refreshData: fetchData, setTracker }}>
            {children}
        </TrackerDataContext.Provider>
    );
}
