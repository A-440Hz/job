import { useState, useEffect } from "react";
import { useTrackerData } from "./JobAppTrackerDataContext";
import { useScreenSize } from "./ScreenSizeProvider";
import { openOneLootbox } from "./api/user";
import { getImageURL } from "./api/collectable";

function LootboxPage() {
    const { user, error, refreshData} = useTrackerData();
    const isDesktop = useScreenSize();

    // State management for view switching
    const [view, setView] = useState('inventory'); // 'inventory' | 'opening'
    const [lootboxResult, setLootboxResult] = useState(null);
    const [isLoading, setIsLoading] = useState(false);

    if (error) return <div>Error loading backend: {error}</div>;
    if ( !user ) return <div>???</div>;

    const handleOpenLootbox = async () => {
        setIsLoading(true);
        try {
            const result = await openOneLootbox();
            setLootboxResult(result);
            setView('opening');
        } catch (error: any) {
            console.error('Error opening lootbox:', error.message);
        }
        setIsLoading(false);
    };

    const handleBackToInventory = () => {
        setView('inventory');
        refreshData(); // Refresh user data to update lootbox count
        setLootboxResult(null);
    };
    

    // Conditional rendering based on view state
    if (view === 'opening') {
        return <OpeningAnimationView result={lootboxResult} onComplete={handleBackToInventory} />;
    }

    // Inventory view
    return (
        <div className="px-8 pt-4 w-8/10 justify-self-center justify-items-center border-blue-200 border mt-3">
            <div className="bg-slate-50 rounded-xl p-6 shadow-md border border-gray-100 mt-4 max-w-md mx-auto">
                <div className="text-center select-none">
                    <div className="text-4xl font-bold text-yellow-600 mb-2">
                        {user?.inventory?.NumLootboxes || 0}
                    </div>
                    <div className="text-sm text-gray-600 mb-6">
                        {user?.inventory?.NumLootboxes === 1 ? 'Lootbox Available' : 'Lootboxes Available'}
                    </div>

                    <div className="flex justify-evenly gap-4">
                        <button
                            className={`px-6 py-3 rounded-lg font-semibold transition-all duration-200 ${
                                (user?.inventory?.NumLootboxes || 0) >= 1 && !isLoading
                                    ? 'bg-yellow-500 text-white hover:bg-yellow-600 shadow-md hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-yellow-500 focus:ring-offset-2'
                                    : 'bg-gray-300 text-gray-500 cursor-not-allowed'
                            }`}
                            disabled={(user?.inventory?.NumLootboxes || 0) < 1 || isLoading}
                            onClick={handleOpenLootbox}
                        >
                            {isLoading ? 'Opening...' : 'Open 1'}
                        </button>

                        <button
                            className={`px-6 py-3 rounded-lg font-semibold transition-all duration-200 ${
                                (user?.inventory?.NumLootboxes || 0) >= 10 && !isLoading
                                    ? 'bg-yellow-500 text-white hover:bg-yellow-600 shadow-md hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-yellow-500 focus:ring-offset-2'
                                    : 'bg-gray-300 text-gray-500 cursor-not-allowed'
                            }`}
                            disabled={(user?.inventory?.NumLootboxes || 0) < 10 || isLoading}
                            onClick={() => {
                                // TODO: Handle opening 10 lootboxes
                                console.log('Opening 10 lootboxes');
                            }}
                        >
                            Open 10
                        </button>
                    </div>
                </div>
            </div>

            <div>
                {user && user.inventory ? JSON.stringify(user.inventory) : 'empty'}
            </div>
        </div>
    )
}

function OpeningAnimationView({ result, onComplete }: { result: any, onComplete: () => void }) {
    const [animationState, setAnimationState] = useState('waiting'); // 'waiting' | 'animating' | 'complete'
    console.log(result);
    const handleInteraction = () => {
        if (animationState === 'waiting') {
            setAnimationState('animating');
            // Simulate animation duration
            setTimeout(() => {
                setAnimationState('complete');
            }, 2000);
        } else if (animationState === 'complete') {
            onComplete();
        }
    };

    return (
        <div className="px-8 pt-4 w-8/10 justify-self-center justify-items-center border-blue-200 border mt-3">
            <div className="bg-white rounded-xl p-8 shadow-lg border border-gray-100 mt-4 max-w-lg mx-auto">
                <div className="text-center">
                    {animationState === 'waiting' && (
                        <div
                            className="cursor-pointer select-none"
                            onClick={handleInteraction}
                            onMouseEnter={handleInteraction}
                        >
                            <h2 className="text-2xl font-bold text-gray-800 mb-6">Ready to Open!</h2>
                            <div className="text-6xl mb-4">📦</div>
                            <p className="text-gray-600">Click or hover to open your lootbox</p>
                        </div>
                    )}

                    {animationState === 'animating' && (
                        <div className="select-none">
                            <h2 className="text-2xl font-bold text-gray-800 mb-6">Opening...</h2>
                            <div className="animate-bounce text-6xl mb-4">📦</div>
                            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-yellow-500 mx-auto"></div>
                        </div>
                    )}

                    {animationState === 'complete' && result && (
                        <div
                            className="cursor-pointer select-none"
                            onClick={handleInteraction}
                        >
                            <h2 className="text-2xl font-bold text-green-600 mb-6">🎉 Lootbox Opened!</h2>

                            {result.col && result.col.Collectable.Filename && (
                                <div className="mb-6">
                                    <img
                                        src={getImageURL(result.col.Collectable.Filename)}
                                        alt="Collectable"
                                        className="w-32 h-32 mx-auto rounded-lg shadow-md object-cover"
                                        onError={(e) => {
                                            e.currentTarget.src = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" width="128" height="128" viewBox="0 0 24 24" fill="%23999"><rect width="24" height="24" fill="%23f5f5f5"/><text x="12" y="12" text-anchor="middle" dy=".3em" fill="%23999">?</text></svg>';
                                        }}
                                    />
                                </div>
                            )}

                            <div className="bg-yellow-50 rounded-lg p-4 mb-6">
                                <h3 className="font-semibold text-gray-800 mb-2">You received:</h3>
                                <div className="text-sm text-gray-600">
                                    {result.col ? (
                                        <div>
                                            <p><strong>Name:</strong> {result.col.Collectable.Name || 'Unknown Item'}</p>
                                            <p><strong>Description:</strong> {result.col.Collectable.Description || 'A rare squid'}</p>
                                        </div>
                                    ) : (
                                        <p>Mysterious item received!</p>
                                    )}
                                </div>
                            </div>

                            <p className="text-gray-500 text-sm">Click anywhere to return to inventory</p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

export default LootboxPage;