import React, { createContext, useContext, useEffect, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchCollectablesData } from './api/user';
import { fetchTrackerData } from './api/tracker';
import { cacheCollectables } from './utils/cacheCollectables';

// NOTE: the CollectablesDataContext differs from TrackerDataContext in that it
// specifically handles collectables-related data and caching.
// To this end, it contains a no-cache header on fetches to ensure fresh data retrieval.

type CollectablesDataContextType = {
    user: any;
    earned_collectables: any[];
    all_collectables: any[];
    error: string | null,
    refreshData: () => void;
};

const CollectablesDataContext = createContext<CollectablesDataContextType>({
    user: null,
    earned_collectables: [],
    all_collectables: [],
    error: null,
    refreshData: () => {},
});

export function useCollectablesData() {
    return useContext(CollectablesDataContext);
};

export function CollectablesDataProvider({ children }: { children: React.ReactNode }) {
    // Custom query function with retry logic that generates a new session if needed
    const fetchCollectablesWithRetry = async () => {
        try {
            const data = await fetchCollectablesData();
            // Request service worker to cache collectable images
            await cacheCollectables(data.all_collectables);
            return data;
        } catch (err) {
            // If fetch fails, try to generate a new session and retry
            await fetchTrackerData();
            const data = await fetchCollectablesData();
            await cacheCollectables(data.all_collectables);
            return data;
        }
    };

    // Use React Query to fetch and cache collectables data
    const { data, error, refetch } = useQuery({
        queryKey: ['collectablesData'],
        queryFn: fetchCollectablesWithRetry,
    });

    return (
        <CollectablesDataContext.Provider
            value={{
                user: data?.user || null,
                earned_collectables: data?.earned_collectables || [],
                all_collectables: data?.all_collectables || [],
                error: error?.message || null,
                refreshData: refetch,
            }}
        >
            {children}
        </CollectablesDataContext.Provider>
    );
};

export function useInView(options?: IntersectionObserverInit): [React.RefObject<HTMLDivElement | null>, boolean] {
    const ref = useRef<HTMLDivElement>(null);
    const [inView, setInView] = useState(false);

    useEffect(() => {
        const observer = new window.IntersectionObserver(([entry]) => {
            setInView(entry.isIntersecting);
        }, options);

        if (ref.current) observer.observe(ref.current);

        return () => {
            if (ref.current) observer.unobserve(ref.current);
        };
    }, [options]);

    return [ref, inView];
}
