import { useState, useEffect } from "react";
import { useTrackerData } from "./JobAppTrackerDataContext";



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

    const created = new Date(tracker.CreatedAt).toDateString();
    const firstCompleted = new Date(tracker.FirstCompleted).toDateString();
    const lastCompleted = new Date(tracker.LastCompleted).toDateString();

    return (
        <>
        <div className="justify-self-center grid content-center text-center">
            {user.Registered === false && showBanner && (
                <div onClick={() => setShowBanner(false)} className="cursor-pointer">
                <p className='flex left-0 right-0 bg-amber-200 opacity-80 marquee border-b border-amber-400 hover:opacity-50 transition-opacity'> 
                    <span className="text-s text-nowrap text-black p-1 rounded-lg select-none">
                    You are currently using a demo account. Your progress will be lost after 60 days of inactivity or if you lose your session cookie. Register an account to persist your progress
                    </span>
                </p>
                </div>
            )}

        <div className="pt-10 border-2 min-w-[30vw]">
            <p> Username: {user.Username} </p>
            <p> Email: {user.Email} </p>
            <p> Current goal streak: {tracker.CurGoalStreak} </p>
            <p> Longest goal streak: {tracker.MaxGoalStreak} </p>
            <p> Longest daily streak: {tracker.MaxDailyStreak} </p>
            <p> Total applications logged: {tracker.TotalItemsCompleted} </p>
            <p> Max applications within one deadline: {tracker.MaxCycleItemsCompleted} </p>
            <p> Total Lootboxes Earned: {tracker.TotalBoxesAwarded} </p>
            <p> Account Created: {created} </p>
            <p> First application logged: {firstCompleted} </p>
            <p> Last application logged: {lastCompleted} </p>        
        </div>
        </div>
        </>
    )
}

export default UserPage