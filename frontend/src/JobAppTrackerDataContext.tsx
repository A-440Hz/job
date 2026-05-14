import React, { createContext, useContext, useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchTrackerData } from './api/tracker';
import { queryClient } from './queryClient';

/**
 * Type definition for the Tracker Data Context
 * @typedef {Object} TrackerDataContextType
 * @property {Object} user - The current user data
 * @property {Object} tracker - The job application tracker data
 * @property {Date} serverDay - The current server date
 * @property {string|null} error - Any error message from data fetching
 * @property {Function} refreshData - Function to refresh the tracker data
 * @property {Function} setTracker - Function to update the tracker data in cache
 * @property {string} modelName - The name of the selected LLM model
 * @property {Function} setModelName - Function to update the selected LLM model name
 */
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

/**
 * Provides the Job Application Tracker Data Context to child components.
 * @param {React.ReactNode} children - The children to render within the context provider.
 * @returns The context provider component.
 */
export function TrackerDataProvider({ children }: { children: React.ReactNode }) {
    const DEFAULT_MODEL = "openrouter/free";

    // Load model preference from localStorage, or use default
    const [modelName, setModelName] = useState<string>(() => {
        const savedModel = localStorage.getItem('selectedModel');
        return savedModel || DEFAULT_MODEL;
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
