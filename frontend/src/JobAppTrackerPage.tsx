// import { useEffect, useState } from 'react'
// import { fetchTrackerData } from './api/tracker'
import { useTrackerData } from './JobAppTrackerDataContext';
import './App.css'

function JobAppTrackerPage() {
  // const [count, setCount] = useState(0)
  // const [data, setData] = useState<any>(null);
  // const [error, setError] = useState<string | null>(null);

  // useEffect(() => {
  //   useTrackerData()
  //     .then(setData)
  //     // .catch((err) => setError(err.message));
  // }, []);

  const { data, error, refreshData } = useTrackerData();
  if (error) return <div>Error: {error}</div>;
  if (!data) return <div></div>;

  // if (error) return <p className="text-red-500">Error: {error}</p>;
  // if (!data) return <p>Loading...</p>;

  return (
    <div className="p-8 w-8/10 justify-self-center border-blue-200 border mt-20">
      <h1 className="text-2xl font-bold text-indigo-700 mb-4 justify-self-center">
        Hellooo, {data.user?.ID ?? 'user'}!
      </h1>
      <h2 className="text-xl mb-2">Your Job Applications:</h2>
      <ul className="space-y-2">
        {data.tracker?.Items?.map((item: any) => (
          <li key={item.ID} className="p-4 rounded bg-white shadow">
            <p className="font-semibold text-blue-600">{item.Title}</p>
            <p className="text-sm text-gray-500">{item.Body}</p>
          </li>
        ))}
      </ul>
    </div>
  );

}

export default JobAppTrackerPage
