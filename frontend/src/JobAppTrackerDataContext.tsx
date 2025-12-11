import React, { createContext, useContext, useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchTrackerData } from './api/tracker';
import { queryClient } from './queryClient';

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
    // Load model preference from localStorage, or use default
    const [modelName, setModelName] = useState<string>(() => {
        const savedModel = localStorage.getItem('selectedModel');
        return savedModel || "tngtech/deepseek-r1t2-chimera:free";
    });

    // Save model preference to localStorage whenever it changes
    useEffect(() => {
        localStorage.setItem('selectedModel', modelName);
    }, [modelName]);

    // Use React Query to fetch and cache tracker data
    const { data, error, refetch } = useQuery({
        queryKey: ['trackerData'],
        queryFn: fetchTrackerData,
    });

    // Function to update tracker in cache without refetching
    const setTracker = (newTracker: any) => {
        queryClient.setQueryData(['trackerData'], (oldData: any) => {
            if (!oldData) return oldData;
            return {
                ...oldData,
                tracker: newTracker,
            };
        });
    };

    return (
        <TrackerDataContext.Provider value={{
            user: data?.user || null,
            tracker: data?.tracker || null,
            serverDay: data?.serverDay ? new Date(data.serverDay) : undefined,
            error: error?.message || null,
            refreshData: refetch,
            setTracker,
            modelName,
            setModelName,
        }}>
            {children}
        </TrackerDataContext.Provider>
    );
}
