import './Card.css'

export default function Card({ accountInfo, theme = {
  backgroundColor: "blue",
  color: "white",
}, type = "debit" }) {

  // common styling of all cards  
  const commonStyle = {
    width: "280px",
    minWidth: "280px", // make sure it's fixed
    height: "180px",
    borderRadius: "15px",
    padding: "10px",
    ...theme
  };
  const top = type === "debit" ? 80 : 40;

  return (
    <div className="account-card" style={commonStyle}>
      <div className="institution-name" style={{fontSize: "20px"}}>
        {accountInfo.institution}
      </div>
      <div className="account-name" style={{fontSize: "16px"}}>
        {accountInfo.name}
      </div>
      <div className="account-num" style={{ position: "relative", top: "120px" }}>
        ****{accountInfo.account_number}
      </div>
      <div style={{float: "right"}}>
        {
          type === "credit" && (
            <div className="credit-limit" style={{fontSize: "18px"}}>
              <span style={{display: "block", fontSize: "13px"}}>Credit limit: </span>
              ${accountInfo.credit_limit}
            </div>
        )}
        <div className="balance" style={{ 
          fontSize: "25px", 
          position: "relative", 
          top: `${top}px`
        }}>
          <span style={{ display: "block", fontSize: "13px" }}>Balance: </span>
          ${accountInfo.balance}
        </div>
      </div>
    </div>
  );
}