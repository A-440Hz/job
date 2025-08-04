import { useTrackerData } from "./JobAppTrackerDataContext";



function UserPage() {
    const { user, error} = useTrackerData();
    if (error) return <div>Error loading backend: {error}</div>;
    if ( !user ) return <div>???</div>;

    return (
        <div className="justify-self-center">
        <p> {user.ID} </p>
        <p> {user.Registered? "registered" : "not registered"} </p>
        <p> {user.CreatedAt} </p>
        <p> {user.UpdatedAt} </p>
        <p> {user.DeletedAt} </p>
        </div>
    )
}

export default UserPage