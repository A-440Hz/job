import { useState } from 'react';
import { useTrackerData } from './JobAppTrackerDataContext';
import { NavLink } from 'react-router-dom';

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

export function Navbar() {
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
      Open Lootbox
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

export function Login() {
  return <a className="text-text-secondary pr-1.5 md:pr-3 hover:opacity-60" id="navbar_sign_in_button" href="/login">Login</a>
}

export default function Topbar() {
  const {user, error} = useTrackerData();
  if (error) return <div>Error loading backend: {error}</div>;
  if (!user) return <div>???</div>; 
    return (
    <div className='flex h-[62px] min-w-full items-center border justify-between select-none z50'>
      <span> <Glyph /> </span>
      {/* <span className="hidden md:flex border select-none text-4xl ml-0"> Hi this is a topbar </span> */}
      <Navbar />
      <Login />
      {/* TODO: scrolling banner; click to hide */}
      <p className='flex absolute top-[60px] right-0 text-xs text-nowrap'> 
        {user.Registered? '' : 'You are logged in as a demo user. Your progress will be lost after 60 days of inactivity or if you lose your session cookie. Register an account to persist your progress'}
      </p>
    </div>);
}

