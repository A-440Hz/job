import { useState, useEffect } from "react";
import { useCollectablesData, useInView } from "./CollectablesDataContext";
import { useScreenSize } from "./ScreenSizeProvider";
import { getImageURL, filenameToTitle } from "./api/collectable";
import ReactPlayer from "react-player";

const rarityMap: Map<string, string> = new Map([
    ["C", "Common"],
    ["B", "Rare"],
    ["A", "Epic"],
    ["S", "Legendary"],
]);

export function valueToRarity(v: string): string {
    // TODO: can return color and styled element
    return rarityMap.get(v) || "Unknown";
}

export function CollectableMedia({ collectable, onClick, lootboxView }: { collectable: any, onClick?: (e: React.MouseEvent) => void, lootboxView?: boolean }) {
    const [containerRef, inView] = useInView({ threshold: 0.1 });
    const [isPlaying, setIsPlaying] = useState(false);

    if (!collectable) return null;

    return (
        <div ref={containerRef} 
            className="w-32 h-32 mx-auto rounded-lg shadow-md object-cover bg-gray-100 cursor-pointer"
            onMouseOver={() => setIsPlaying(true)}
            onMouseLeave={() => setIsPlaying(false)}
        >
            {(inView || lootboxView) && collectable.Type === "media" ? (
                <ReactPlayer
                    src={getImageURL(collectable.Filename)}
                    playing={lootboxView || isPlaying}
                    loop={true}
                    controls={false}
                    muted={true}
                    playsInline={true}
                    style={{
                        width: "100%",
                        height: "100%",
                    }}
                    onClick={onClick}
                    className="mx-auto rounded-lg shadow-md object-cover"
                />
            ) : null}
            {(inView || lootboxView) && collectable.Type !== "media" ? (
                <img
                    src={getImageURL(collectable.Filename)}
                    alt={filenameToTitle(collectable.Name)}
                    className="w-32 h-32 mx-auto rounded-lg shadow-md object-cover"
                    onClick={onClick}
                    onError={(e) => {
                        e.currentTarget.src = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" width="128" height="128" viewBox="0 0 24 24" fill="%23999"><rect width="24" height="24" fill="%23f5f5f5"/><text x="12" y="12" text-anchor="middle" dy=".3em" fill="%23999">?</text></svg>';
                    }}
                />
            ) : null}
        </div>
    );
        
    }
    
export const MagnifiedMediaModal = ( {showMagnified, setShowMagnified, collectable}: {showMagnified: boolean, setShowMagnified: (show: boolean) => void, collectable: any} ) => {
    if (!showMagnified || !collectable) return null;

    if (collectable.Type === "media") {
        return (
        <div
            className="fixed inset-0 bg-black bg-opacity-60 flex items-center justify-center z-50 select-none"
            onClick={() => setShowMagnified(false)}
        >
            <div className="w-full flex flex-col items-center mb-4">
                <h1 className="text-white text-4xl font-bold text-center mb-4">
                    {filenameToTitle(collectable.Name) || 'A Rare Squid'}
                </h1>
                    <ReactPlayer
                        src={getImageURL(collectable.Filename)}
                        playing={true}
                        loop={true}
                        controls={false}
                        muted={true}
                        playsInline={true}
                        style={{
                            width: "100%",
                            height: "100%",
                        }}
                        className="rounded-xl border-4 border-yellow-400"
                    />
                    <div className="text-white text-center mt-4 px-4">
                    {/* {"text"} */}
                    </div>
            </div>
        </div>
        );
    }
    return (
        <div
            className="fixed inset-0 bg-black bg-opacity-60 flex items-center justify-center z-50 select-none"
            onClick={() => setShowMagnified(false)}
        >
            <div className="w-full flex flex-col items-center mb-4">
                <h1 className="text-white text-4xl font-bold text-center mb-4">
                    {filenameToTitle(collectable.Name) || 'A Rare Squid'}
                </h1>
                <img
                    src={getImageURL(collectable.Filename)}
                    alt={filenameToTitle(collectable.Name) || 'A Rare Squid'}
                    className="w-[32rem] h-[32rem] rounded-xl shadow-2xl object-contain border-4 border-yellow-400"
                    style={{ maxWidth: '90vw', maxHeight: '90vh' }}
                    loading="lazy"
                />
                <div className="text-white text-center mt-4 px-4">
                    {/* {"text"} */}
                </div>
            </div>
        </div>
    );
};

export function unearnedCollectable() {
    // return a blank card with a question mark
    return {}

}

export function CollectablesPage() {
    const { user, earned_collectables, all_collectables, error, refreshData } = useCollectablesData();
    const isDesktop = useScreenSize();

    const [view, setView] = useState('viewEarned'); // 'viewEarned' | 'viewAll'
    const [showMagnified, setShowMagnified] = useState(false);
    const [selectedCollectable, setSelectedCollectable] = useState(null);

    useEffect(() => {
        if (user) {
        refreshData();
        }
    }, []);

    if (error) return <div>Error loading backend: {error}</div>;
    if ( !user || !earned_collectables || !all_collectables ) return <div>???</div>;

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
                    onClick={() => setView('viewEarned')}
                >
                    View Earned
                </button>
                <button
                    className={`px-6 py-3 rounded-lg font-semibold transition-all duration-200 ${
                        view === 'viewAll'
                            ? 'bg-yellow-500 text-white shadow-md hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-yellow-500 focus:ring-offset-2'
                            : 'bg-gray-300 text-gray-500 hover:bg-yellow-400 hover:text-white cursor-pointer'
                    }`}
                    onClick={() => setView('viewAll')}
                >
                    View All
                </button>
            </div>

            {view === 'viewEarned' && (
                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-6 p-4">
                    {earned_collectables.map((item, index) => (
                        <div key={item.ID || index} className="flex flex-col items-center">
                            <div className="cursor-pointer transform transition-transform hover:scale-105 w-32 rounded-lg overflow-hidden bg-gray-100">
                                <CollectableMedia
                                    collectable={item.Collectable}
                                    onClick={() => handleCollectableClick(item.Collectable)}
                                    lootboxView={false}
                                />
                            </div>
                            <div className="text-center mt-2">
                                <p className="text-xs font-medium text-yellow-700 truncate max-w-full">
                                    {filenameToTitle(item.Collectable?.Name) || 'Unknown'}
                                </p>
                                <p className="text-xs text-gray-500 mt-0.75">{valueToRarity(item.Collectable?.Value)}</p>
                                <p className="text-xs text-gray-500 mt-0.5">
                                    Quantity: {item.Quantity}
                                </p>
                            </div>
                        </div>
                    ))}
                </div>
            )}

            {view === 'viewAll' && (
                <div className="text-center text-gray-500 py-8">
                    <p>View All collectables coming soon...</p>
                </div>
            )}
        </div>
    )

}
