import React, { createContext, useContext, useEffect, useState } from 'react';
import { fetchTrackerData } from './api/tracker';

type TrackerDataContextType = {
    user: any;
    tracker: any;
    serverDay: any;
    error: string | null;
    refreshData: () => void;
    setTracker: (d: any) => void;
};

const TrackerDataContext = createContext<TrackerDataContextType>({
    user: null,
    tracker: null,
    serverDay: Date,
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
    const [serverDay, setServerDay] = useState<Date>();
    const [error, setError] = useState<string | null>(null);

    const fetchData = () => {
        fetchTrackerData()
            .then((data) => {
                setUser(data.user);
                setTracker(data.tracker);
                setServerDay(new Date(data.serverDay));
            })
            .catch((err) => setError(err.message));
    };

    useEffect(fetchData, []);

    return (
        <TrackerDataContext.Provider value={{ user, tracker, serverDay, error, refreshData: fetchData, setTracker }}>
            {children}
        </TrackerDataContext.Provider>
    );
}
