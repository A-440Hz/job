import React, { createContext, useContext, useEffect, useState } from 'react';
import { fetchCollectablesData } from './api/user';
import { fetchTrackerData } from './api/tracker';

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
        } catch (err) {
            try {
                await fetchTrackerData();
                const data = await fetchCollectablesData();
                setUser(data.user);
                setEarnedCollectables(data.earned_collectables);
                setAllCollectables(data.all_collectables);
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