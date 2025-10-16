import { useEffect, useState } from 'react';
import { useTrackerData } from './JobAppTrackerDataContext';
import { logoutUser } from './api/user';
import { NavLink } from 'react-router-dom';
import { useScreenSize } from './ScreenSizeProvider';

function Glyph() {
  const [source, setSource] = useState("/src/assets/sb1rb.png");
  const handleMouseAway: React.MouseEventHandler<HTMLImageElement> = (event) => {
    event.preventDefault();
    setSource("/src/assets/sb2rb.png");
  };
  const handleMouseOver: React.MouseEventHandler<HTMLImageElement> = (event) => {
    event.preventDefault();
    setSource("/src/assets/sb1rb.png");
  };
  const handleMouseClick = () => {
    window.location.href = "/";
  }
  return <img className="flex select-none min-w-12 max-w-24 mr-0 hover:cursor-pointer" id="glyph" 
    onMouseEnter={handleMouseAway} 
    onMouseLeave={handleMouseOver}
    onClick={handleMouseClick} 
    src={source}
    draggable="false"
  /> 
}

export function Navbar({ user }: { user: any }) {
  const navClass = "mx-1.5 text-m/6 hover:opacity-60 "
  return <nav className='flex border-2 justify-between pl-3 pr-3'>
    <NavLink to='/' className={({ isActive }) =>
        isActive ? navClass + "text-amber-300" : navClass
      }>
      Tracker
    </NavLink>
    <NavLink to='/Collection' className={({ isActive }) =>
        isActive ? navClass+ "text-amber-300" : navClass
      }>
      Collection
    </NavLink>
    <NavLink to='/Lootbox' className={({ isActive }) =>
        isActive ? navClass + "text-amber-300" : navClass
      }>
       Lootbox <text className="text-orange-300"> {(user.inventory?.NumLootboxes > 0)? `(${user.inventory?.NumLootboxes})` : ''}</text>
    </NavLink>
    <NavLink to='/About' className={({ isActive }) =>
        isActive ? navClass + "text-amber-300" : navClass
      }>
      About
    </NavLink>
    <NavLink to='/profile' className={({ isActive }) =>
        isActive ? navClass + "text-amber-300" : navClass
      }>
      Profile
    </NavLink>
  </nav>;
} 

export function Login({ user }: { user: any }) {
  if (!user) return <div>???</div>;
  if (user.Registered === true) {
    return <a className="text-text-secondary pr-1.5 md:pr-3 hover:opacity-60" id="navbar_sign_in_button" onClick={() => logoutUser()}>Logout</a>
  } else {
    return <a className="text-text-secondary pr-1.5 md:pr-3 hover:opacity-60" id="navbar_sign_in_button" href="/login">Login/Register</a>
  }
}

export default function Topbar() {
  // const isDesktop = useScreenSize();
  const {user, error, refreshData} = useTrackerData();
  useEffect(() => {
    refreshData();
  }, []);

  if (error) return <div>Error loading backend: {error}</div>;
  if (!user) return <div>???</div>; 
    return (
    <div className='flex h-[62px] min-w-full items-center border justify-between select-none z50'>
      <span> <Glyph /> </span>
      {/* <span className="md:flex select-none font-semibold text-3xl ml-0"> {isDesktop && "haotianswebsite.com"} </span> */}
      <Navbar user={user}/>
      <Login user={user} />
    </div>);
}

