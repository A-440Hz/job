import React, { createContext, useContext, useEffect, useRef, useState } from 'react';
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
    const [user, setUser] = useState<any>(null);
    const [earned_collectables, setEarnedCollectables] = useState<any[]>([]);
    const [all_collectables, setAllCollectables] = useState<any[]>([]);
    const [error, setError] = useState<string | null>(null);
    

    const fetchData = async () => {
        setError(null);
        try {
            const data = await fetchCollectablesData();
            setUser(data.user);
            setEarnedCollectables(data.earned_collectables);
            setAllCollectables(data.all_collectables);

            // Request service worker to cache collectable images
            await cacheCollectables(data.all_collectables);
        } catch (err) {
            try {
                await fetchTrackerData(); // this line attempts to generate a new session in case the current one is invalid
                const data = await fetchCollectablesData();
                setUser(data.user);
                setEarnedCollectables(data.earned_collectables);
                setAllCollectables(data.all_collectables);

                // Request service worker to cache collectable images
                await cacheCollectables(data.all_collectables);
            } catch (finalErr: any) {
                setError(finalErr.message || "Failed to load collectables.");
            }
        }
    };

    useEffect(() => {
        fetchData();
    }, []);
    return (
        <CollectablesDataContext.Provider
            value={{
                user,
                earned_collectables,
                all_collectables,
                error,
                refreshData: fetchData,
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
