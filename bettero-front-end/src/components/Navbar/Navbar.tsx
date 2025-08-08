import { useState } from 'react';
import { Link } from 'react-router-dom';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faHouse, faChartSimple, faMoneyBill, faArrowTrendUp } from '@fortawesome/free-solid-svg-icons'
import LogoutForm from '@Forms/DeleteForms/LogoutForm';
import './Navbar.css';


function Navbar() {
  const [logoutFormPresent, setLogoutFormPresent] = useState(false);

  // the content of the header
  const commonStyle = {
    textDecoration: "none",
    color: "white",
    display: "inline-block",
    margin: "0 10%",
    borderRadius: "10px",
  }

  return (
    <>
      <div className="navbar">
        <header>
          <span
            style={{ display: "inline-block", borderRadius: "10px", width: "100%" }}
            onMouseOver={(e) => e.currentTarget.style.backgroundColor = "rgba(0, 0, 0, 0.25)"}
            onMouseOut={(e) => e.currentTarget.style.backgroundColor = "darkcyan"}>
            Bettero App
          </span>
        </header>
        <div className="options-wrapper">
          <Link to="/" 
            style={commonStyle}
            onMouseOver={(e) => e.currentTarget.style.backgroundColor = "rgba(0, 0, 0, 0.25)"}
            onMouseOut={(e) => e.currentTarget.style.backgroundColor = "darkcyan"}
          >
            <FontAwesomeIcon icon={faHouse} style={{marginRight: "5px"}} />Main
          </Link>
          <Link to="/summary" 
            style={commonStyle}
            onMouseOver={(e) => e.currentTarget.style.backgroundColor = "rgba(0, 0, 0, 0.25)"}
            onMouseOut={(e) => e.currentTarget.style.backgroundColor = "darkcyan"}
          >
            <FontAwesomeIcon icon={faChartSimple} style={{marginRight: "5px"}} />Summary
          </Link>
          <Link to="/budget" 
            style={commonStyle}
            onMouseOver={(e) => e.currentTarget.style.backgroundColor = "rgba(0, 0, 0, 0.25)"}
            onMouseOut={(e) => e.currentTarget.style.backgroundColor = "darkcyan"}
          >
            <FontAwesomeIcon icon={faMoneyBill} style={{marginRight: "5px"}} />Budget
          </Link>
          <Link to="/investment" 
            style={commonStyle}
            onMouseOver={(e) => e.currentTarget.style.backgroundColor = "rgba(0, 0, 0, 0.25)"}
            onMouseOut={(e) => e.currentTarget.style.backgroundColor = "darkcyan"}
          >
            <FontAwesomeIcon icon={faArrowTrendUp} style={{marginRight: "5px"}} />Investment
          </Link>
          <div 
            style={commonStyle}
            onMouseOver={(e) => e.currentTarget.style.backgroundColor = "rgba(0, 0, 0, 0.25)"}
            onMouseOut={(e) => e.currentTarget.style.backgroundColor = "darkcyan"}
            onClick={() => setLogoutFormPresent(!logoutFormPresent)}
          >
            LOG OUT
          </div>
        </div>
      </div>
      {logoutFormPresent && <LogoutForm
        onHide={() => setLogoutFormPresent(!logoutFormPresent)} />}
    </>
  );
}

export default Navbar; 
