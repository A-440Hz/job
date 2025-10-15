import { useState, useEffect } from "react";
import { CollectableMedia, MagnifiedMediaModal, valueToRarity } from "./Collectables"
import { useCollectablesData } from "./CollectablesDataContext";
import { useScreenSize } from "./ScreenSizeProvider";
import { openOneLootbox } from "./api/user";
import { getImageURL, filenameToTitle } from "./api/collectable";


// It's strange to press the back button and see a lootbox count that is inaccurate.
// When the count went from 0 to 1, the button is enabled but unpressable. 
//
// Making a fetch request within `useEffect` on every render is generally not recommended due to performance concerns and potential infinite loops. However, there are strategies to keep the frontend state fresh without constantly fetching data. 
// ## Pros and Cons of Fetching on Every Render
// | **Pros**                   | **Cons**                     |
// |---------------------------|------------------------------|
// | Always up-to-date data    | High resource usage          |
// | Simple to implement        | Risks infinite loop          |
// | Useful for debugging       | Can lead to poor user experience due to lag |
// ---
// ## Recommended Solutions to Keep State Fresh
// ### 1. **Dependency Array in `useEffect`**
//    - You can control when the fetch is executed by passing a dependency array to `useEffect`. By including the state or props that trigger changes, you can fetch data only when those values change.
// ### 2. **Polling**
//    - Set up a timer using `setInterval` inside `useEffect` to fetch data every few seconds or minutes. Make sure to clear the interval on component unmount.
// ### 3. **WebSockets**
//    - For real-time applications, consider using WebSockets. This allows the server to push updates to the client, maintaining a fresh state without continuous polling.
// ### 4. **State Management Libraries**
//    - Use libraries like Redux or Zustand that handle global state and can integrate with middlewares (like Redux Thunk) to manage side effects, including fetching data.
// ### 5. **React Query**
//    - This library simplifies data fetching, caching, and synchronization. It offers features like automatic refetching, background updates, and caching.
// ### 6. **Manual Refresh**
//    - Provide a button or user-triggered event that allows users to refresh data as needed. This gives control without unnecessary requests.
// ### 7. **Conditional Fetching**
//    - You can conditionally fetch data based on specific user actions or conditions that make a fetch specifically necessary.
// By combining these strategies, you can effectively keep your frontend state fresh without compromising performance or user experience. Would you like to dive deeper into any of these solutions?

function LootboxPage() {
    const { user, error, refreshData} = useCollectablesData();
    const isDesktop = useScreenSize();

    // State management for view switching
    const [view, setView] = useState('inventory'); // 'inventory' | 'opening'
    const [lootboxResult, setLootboxResult] = useState(null);
    const [isLoading, setIsLoading] = useState(false);

    useEffect(() => {
        refreshData();
    }, []);
    
    useEffect(() => {
        if (user) {
        refreshData();
        }
    }, []);

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
        <div className="mx-auto px-8 pt-4 max-w-8/10 justify-self-center justify-items-center border-blue-200 border mt-3">
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
    const [showMagnified, setShowMagnified] = useState(false);
    const [isPreloaded, setIsPreloaded] = useState(false);

    // Preload the collectable img when component mounts
    useEffect(() => {
        if (result?.col?.Collectable?.Filename && result.col.Collectable.Type === "image") {
            const preloadElement = document.createElement('img');

            preloadElement.onload = () => setIsPreloaded(true);
            preloadElement.src = getImageURL(result.col.Collectable.Filename);
        }
    }, [result]);

    const handleInteraction = () => {
        if (animationState === 'waiting') {
            setAnimationState('animating');
            // Simulate animation duration
            setTimeout(() => {
                setAnimationState('complete');
            }, 600);
        } else if (animationState === 'complete') {
            onComplete();
        }
    };

        

    return (
        <div className="mx-auto px-8 pt-4 w-8/10 justify-self-center justify-items-center border-blue-200 border mt-3">
            <MagnifiedMediaModal showMagnified={showMagnified} setShowMagnified={setShowMagnified} collectable={result.col.Collectable} />
            <div className="bg-white rounded-xl p-8 shadow-lg border border-gray-100 mt-4 max-w-lg mx-auto">
                <div className="text-center">
                    {animationState === 'waiting' && (
                        <div
                            className="cursor-pointer select-none"
                            onClick={handleInteraction}
                        >
                            <h2 className="text-2xl font-bold text-gray-800 mb-6">Ready to Open!</h2>
                            <div className="animate-bounce text-6xl mb-4">📦</div>
                            <p className="text-gray-600">Click to open your lootbox</p>
                            {!isPreloaded && (
                                <div className="mt-4 flex items-center justify-center gap-2">
                                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-400"></div>
                                    <span className="text-xs text-gray-500">Preparing squids...</span>
                                </div>
                            )}
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
                                    <CollectableMedia
                                        collectable={result.col.Collectable}
                                        onClick={(e: React.MouseEvent) => {
                                            e.stopPropagation();
                                            if (!showMagnified) setShowMagnified(true);
                                        }}
                                        lootboxView={true}
                                    />
                                </div>
                            )}

                            <div className="bg-yellow-50 rounded-lg p-4 mb-6">
                                <h3 className="font-semibold text-gray-800 mb-2">You received:</h3>
                                <div className="text-sm text-gray-600">
                                    {result.col ? (
                                        <div>
                                            <p><strong>Name:</strong> {filenameToTitle(result.col.Collectable.Name) || 'Unknown Item'}</p>
                                            <p><strong>Rarity:</strong> {valueToRarity(result.col.Collectable.Value)}</p>
                                            <p><strong>Number Owned:</strong> {String(result.col.Quantity) || 'Uncertain' }</p>
                                        </div>
                                    ) : (
                                        <p>Mysterious item received!</p>
                                    )}
                                </div>
                            </div>

                            <p className="text-gray-500 text-sm">Click here to return to previous screen</p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );    
}





export default LootboxPage;