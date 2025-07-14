import Navbar from './components/Navbar/Navbar.jsx';
import MainPage from './components/Pages/MainPage/MainPage.jsx';
import AccountPage from './components/Pages/AccountPage/AccountPage.jsx';
import SummaryPage from './components/Pages/SummaryPage/SummaryPage.jsx';
import BudgetPage from './components/Pages/BudgetPage/BudgetPage.jsx';
import InvestmentPage from './components/Pages/InvestmentPage/InvestmentPage.jsx';
import LoginPage from './components/Pages/LoginPage/LoginPage.jsx';
import PrivateRoute from './provider/PrivateRoute.jsx';
import LogoutRoute from './provider/LogoutRoute.jsx';
import LogoutForm from './components/Forms/DeleteForms/LogoutForm.jsx';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {faBars, faX, faHouse, faChartSimple, faMoneyBill, faArrowTrendUp } from '@fortawesome/free-solid-svg-icons'
import { useState } from 'react';
import { useMediaQuery } from '@uidotdev/usehooks';
import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import './App.css';

function HeaderBar () {
  const [menuPresent, setMenuPresent] = useState(false);
  const [logoutFormPresent, setLogoutFormPresent] = useState(false);
  const isMobileDevice = useMediaQuery("only screen and (max-width: 750px)");

  const commonStyle = !isMobileDevice ? 
    {
      textDecoration: "none",
      fontSize: "16px",
      color: "white",
      margin: "5px 0",
      padding: "0 5px",
      borderRadius: "10px",
    } : {
      borderTop: "1px solid black",
      textAlign: "center",
      fontSize: "16px",
      color: "white",
      textDecoration: "none",
    };

  const handleMouseOver = (e) => e.target.style.backgroundColor = "rgba(0, 0, 0, 0.25";
  const handleMouseOut = (e) => e.target.style.backgroundColor = "darkcyan";

  return (
    <>
      <div className="header-bar">
        <div className="title-area">
          <button><Link to="/" onMouseOver={handleMouseOver} onMouseOut={handleMouseOut}
            style={{
              padding: "10px 5px",
              borderRadius: "10px",
              textDecoration: "none",
              color: "white",
            }}>Bettero App</Link></button>
          {isMobileDevice && <button onClick={() => {
            if (isMobileDevice) {
              setMenuPresent(!menuPresent);
            }
          }}>
            {menuPresent ?
              <FontAwesomeIcon icon={faX} style={{color: "white", fontSize: "24px"}} /> :
              <FontAwesomeIcon icon={faBars} style={{color: "white", fontSize: "24px"}} />
            }
          </button>}
        </div>
        {(menuPresent || !isMobileDevice) && (
          <div className="menu-wrapper">
            <Link to="/" style={commonStyle} onMouseOver={handleMouseOver} onMouseOut={handleMouseOut}>
              <FontAwesomeIcon style={{ marginRight: "5px" }} icon={faHouse} />Main
            </Link>
            <Link to={`/summary`} style={commonStyle} onMouseOver={handleMouseOver} onMouseOut={handleMouseOut}>
              <FontAwesomeIcon style={{ marginRight: "5px" }} icon={faChartSimple} />Summary
            </Link>
            <Link to={`/budget`} style={commonStyle} onMouseOver={handleMouseOver} onMouseOut={handleMouseOut}>
              <FontAwesomeIcon style={{ marginRight: "5px" }} icon={faMoneyBill} />Budget
            </Link>
            <Link to={`/investment`} style={commonStyle} onMouseOver={handleMouseOver} onMouseOut={handleMouseOut}>
              <FontAwesomeIcon style={{ marginRight: "5px" }} icon={faArrowTrendUp} />Investment
            </Link>
            <div style={commonStyle} onMouseOver={handleMouseOver} onMouseOut={handleMouseOut}
              onClick={() => setLogoutFormPresent(!logoutFormPresent)}>LOG OUT</div>
          </div>
        )}
      </div>
      {logoutFormPresent && <LogoutForm onHideForm={() => setLogoutFormPresent(!logoutFormPresent)} />}
    </>
  );
}


// export the App component 
const App = () => {
  // parameters of the app 
  // state of presence of the navbar
  // the content of the title button 
  const isMediumDevice = useMediaQuery("only screen and (max-width: 800px)");
  let navbarWidth = !isMediumDevice ? 150 : 0;
  let titleHeight = !isMediumDevice ? 0 : 70;

  return (
    <div>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={
            <PrivateRoute>
              {isMediumDevice ? (<HeaderBar />) : (<Navbar />)}
              <MainPage navbarWidth={navbarWidth} titleHeight={titleHeight} />
            </PrivateRoute>
          } />
          <Route path="/accounts/:type" element={
            <PrivateRoute>
              {isMediumDevice ? (<HeaderBar />) : (<Navbar />)}
              <AccountPage navbarWidth={navbarWidth} titleHeight={titleHeight} />
            </PrivateRoute>
          } />
          <Route path="/summary" element={
            <PrivateRoute>
              {isMediumDevice ? (<HeaderBar />) : (<Navbar />)}
              <SummaryPage navbarWidth={navbarWidth} titleHeight={titleHeight} />
            </PrivateRoute>
          } />
          <Route path="/budget" element={
            <PrivateRoute>
              {isMediumDevice ? (<HeaderBar />) : (<Navbar />)}
              <BudgetPage navbarWidth={navbarWidth} titleHeight={titleHeight} />
            </PrivateRoute>
          } />
          <Route path="/investment" element={
            <PrivateRoute>
              {isMediumDevice ? (<HeaderBar />) : (<Navbar />)}
              <InvestmentPage navbarWidth={navbarWidth} titleHeight={titleHeight} />
            </PrivateRoute>
          } />
          <Route path="/login" element={<LoginPage type="LOGIN" />} />
          <Route path="/register" element={<LoginPage type="SIGNUP" />} />
          <Route path="/logout" element={<LogoutRoute />} />
        </Routes>
      </BrowserRouter>
    </div>
  );
}
export default App;