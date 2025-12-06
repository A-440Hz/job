import React, { createContext, useContext, useEffect, useState } from 'react';
import { fetchTrackerData } from './api/tracker';

type TrackerDataContextType = {
    user: any;
    tracker: any;
    serverDay: any;
    error: string | null;
    refreshData: () => void;
    setTracker: (d: any) => void;
    modelName: string;
    setModelName: (m: string) => void;
};

const TrackerDataContext = createContext<TrackerDataContextType>({
    user: null,
    tracker: null,
    serverDay: Date,
    error: null,
    refreshData: () => {},
    setTracker: () => {},
    modelName: "",
    setModelName: () => {},
});

export function useTrackerData() {
    return useContext(TrackerDataContext);
}

export function TrackerDataProvider({ children }: { children: React.ReactNode }) {
    const [user, setUser] = useState<any>(null);
    const [tracker, setTracker] = useState<any>(null);
    const [serverDay, setServerDay] = useState<Date>();
    const [error, setError] = useState<string | null>(null);
    const [modelName, setModelName] = useState<string>("tngtech/deepseek-r1t2-chimera:free");

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
        <TrackerDataContext.Provider value={{ user, tracker, serverDay, error, refreshData: fetchData, setTracker, modelName, setModelName: setModelName }}>
            {children}
        </TrackerDataContext.Provider>
    );
}
