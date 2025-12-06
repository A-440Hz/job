import { useEffect, useState } from "react";
import { useCollectablesData } from "./CollectablesDataContext";
import {useSearchParams} from "react-router-dom";
// import { useScreenSize } from "./ScreenSizeProvider";
import { filenameToTitle } from "./api/collectable";
import { MagnifiedMediaModal, CollectableMedia, valueToRarity, viewAllCollectables, valueToColor } from "./Collectables";

export default function CollectablesPage() {
    const { user, earned_collectables, all_collectables, error, refreshData } = useCollectablesData();
    // const isDesktop = useScreenSize();

    const [showMagnified, setShowMagnified] = useState(false);
    const [selectedCollectable, setSelectedCollectable] = useState(null);
    const [searchParams, setSearchParams] = useSearchParams();

    const view = searchParams.get("view") || "viewEarned"; // "viewEarned" is the default
    const handleSetView = (newView: string) => {
        setSearchParams({ view: newView });
    }

    useEffect(() => {
        if (user) {
        refreshData();
        }
    }, [user?.UserID]);

    if (error) return <div>Error loading backend: {error}</div>;
    if ( !user || !earned_collectables || !all_collectables ) return <div className='text-center justify-self-center'>I'm on the free version</div>;

    const handleCollectableClick = (collectable: any) => {
        setSelectedCollectable(collectable);
        setShowMagnified(true);
    };

    return (
        <div className="px-8 pt-4 w-full max-w-7xl mx-auto">
            <MagnifiedMediaModal
                showMagnified={showMagnified}
                setShowMagnified={setShowMagnified}
                collectable={selectedCollectable}
            />

            <div className="flex justify-center space-x-4 mb-6">
                <button
                    className={`px-6 py-3 rounded-lg font-semibold transition-all duration-200 ${
                        view === 'viewEarned'
                            ? 'bg-yellow-500 text-white shadow-md hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-yellow-500 focus:ring-offset-2'
                            : 'bg-gray-300 text-gray-500 hover:bg-yellow-400 hover:text-white cursor-pointer'
                    }`}
                    onClick={() => handleSetView('viewEarned')}
                >
                    View Earned
                </button>
                <button
                    className={`px-6 py-3 rounded-lg font-semibold transition-all duration-200 ${
                        view === 'viewAll'
                            ? 'bg-yellow-500 text-white shadow-md hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-yellow-500 focus:ring-offset-2'
                            : 'bg-gray-300 text-gray-500 hover:bg-yellow-400 hover:text-white cursor-pointer'
                    }`}
                    onClick={() => handleSetView('viewAll')}
                >
                    View All
                </button>
            </div>

            {view === 'viewEarned' && (
                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-6 p-4">
                    {earned_collectables.sort((a, b) => {
                            return (a.CollectableID || 0) - (b.CollectableID || 0);
                        }).map((item) => (
                        <div key={item.CollectableID} className="flex flex-col items-center">
                            <div className="cursor-pointer transform transition-transform hover:scale-105 w-32 rounded-lg overflow-hidden bg-gray-100">
                                <CollectableMedia
                                    collectable={item.Collectable}
                                    onClick={() => handleCollectableClick(item.Collectable)}
                                    lootboxView={false}
                                />
                            </div>
                            <div className="text-center mt-2">
                                <p className={`text-xs font-medium ${valueToColor(item.Collectable?.Value)} truncate max-w-full`}>
                                    <strong>{filenameToTitle(item.Collectable?.Name) || 'Unknown'}</strong>
                                </p>
                                <p className={`text-xs text-gray-300 mt-0.75`}>{valueToRarity(item.Collectable?.Value)}</p>
                                <p className="text-xs text-gray-300 mt-0.5">
                                    Quantity: {item.Quantity}
                                </p>
                            </div>
                        </div>
                    ))}
                </div>
            )}
            {view === 'viewAll' && (
                viewAllCollectables({earned_collectables, all_collectables, handleCollectableClick})
            )}
        </div>
    )
}