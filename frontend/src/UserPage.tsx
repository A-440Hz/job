import { useState, useEffect } from "react";
import { useTrackerData } from "./JobAppTrackerDataContext";

interface StatRowProps {
    label: string;
    value: string | number;
}

function StatRow({ label, value }: StatRowProps) {
    return (
        <div className="flex justify-between items-center py-3 border-b border-gray-100 last:border-b-0">
            <span className="text-gray-600 font-medium">{label}</span>
            <span className="text-gray-900">{value}</span>
        </div>
    );
}

function UserPage() {
    const {tracker, user, error, refreshData } = useTrackerData();
    if (error) return <div>Error loading backend: {error}</div>;
    if ( !user || !tracker ) return <div>???</div>;

    const [showBanner, setShowBanner] = useState(true);

    useEffect(() => {
        if (user || tracker) {
        refreshData();
        }
    }, [user, tracker]);

    const formatDate = (dateString: string) => {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
    };

    const created = formatDate(tracker.CreatedAt);
    const firstCompleted = formatDate(tracker.FirstCompleted);
    const lastCompleted = formatDate(tracker.LastCompleted);

    return (
        <>
        <div className="justify-self-center grid content-center text-center px-4">
            {user.Registered === false && showBanner && (
                <div onClick={() => setShowBanner(false)} className="cursor-pointer">
                <p className='flex left-0 right-0 bg-amber-200 opacity-80 marquee border-b border-amber-400 hover:opacity-50 transition-opacity'>
                    <span className="text-s text-nowrap text-black p-1 rounded-lg select-none">
                    You are currently using a demo account. Your progress will be lost after 60 days of inactivity or if you lose your session cookie. Register an account to persist your progress
                    </span>
                </p>
                </div>
            )}

        <div className="pt-10 min-w-[30vw] max-w-2xl mx-auto space-y-6">
            {/* User Info Card */}
            <div className="bg-orange-200 rounded-xl shadow-md p-6">
                <h2 className="text-2xl font-bold text-gray-800 mb-4">Account Information</h2>
                <StatRow label="Username" value={user.Username} />
                <StatRow label="Email" value={user.Email} />
                <StatRow label="Account Created" value={created} />
            </div>

            {/* Streaks Card */}
            <div className="bg-orange-200 rounded-xl shadow-md p-6">
                <h2 className="text-2xl font-bold text-gray-800 mb-4">Streaks</h2>
                <StatRow label="Current Goal Streak" value={tracker.CurGoalStreak} />
                <StatRow label="Longest Goal Streak" value={tracker.MaxGoalStreak} />
                <StatRow label="Longest Daily Streak" value={tracker.MaxDailyStreak} />
            </div>

            {/* Activity Stats Card */}
            <div className="bg-orange-200 rounded-xl shadow-md p-6">
                <h2 className="text-2xl font-bold text-gray-800 mb-4">Activity Statistics</h2>
                <StatRow label="Total Applications Logged" value={tracker.TotalItemsCompleted} />
                <StatRow label="Max Applications Within One Deadline" value={tracker.MaxCycleItemsCompleted} />
                <StatRow label="Total Lootboxes Earned" value={tracker.TotalBoxesAwarded} />
                <StatRow label="First Application Logged" value={firstCompleted} />
                <StatRow label="Last Application Logged" value={lastCompleted} />
            </div>
        </div>
        </div>
        </>
    )
}

export default UserPage