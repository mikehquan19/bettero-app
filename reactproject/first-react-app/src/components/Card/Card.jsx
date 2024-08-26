import './Card.css';
import { useMediaQuery } from '@uidotdev/usehooks';

const Card = ({ accountInfo, theme = {
    backgroundColor: "blue",
    color: "white",
}, type = "debit" }) => {
    

    // common styling of all cards  
    let commonStyle = {
        width: "250px",
        minWidth: "250px", // make sure it's fixed
        height: "150px",
        borderRadius: "15px",
        padding: "10px",
    };
    let top = type === "debit" ? 50 : 10;
    // add the properties of the theme object to the common style object
    Object.assign(commonStyle, theme);

    return (
        <div className="account-card" style={commonStyle}>
            <div className="institution-name" style={{
                fontSize: "20px",
            }}>{accountInfo.institution}</div>
            <div className="account-name" style={{
                fontSize: "16px",
            }}>{accountInfo.name}</div>
            <div className="account-num" style={{
                position: "relative",
                top: "90px",
            }}>****{accountInfo.account_number}</div>
            <div style={{
                float: "right",
            }}>
                {
                    // conditional rendering using && operator 
                    // if the account is credit, then shows the credit limit
                    type === "credit" && (
                        <div className="credit-limit" style={{
                            fontSize: "18px",
                        }}>
                            <span style={{
                                display: "block",
                                fontSize: "13px",
                            }}>Credit limit: </span>
                            ${accountInfo.credit_limit}
                        </div>
                    )
                }
                <div className="balance" style={{
                    fontSize: "25px",
                    position: "relative",
                    top: `${top}px`,
                }}>
                    <span style={{
                        display: "block",
                        fontSize: "13px",
                    }}>Balance: </span>
                    ${accountInfo.balance}
                </div>
            </div>
        </div>
    );
}

export default Card;