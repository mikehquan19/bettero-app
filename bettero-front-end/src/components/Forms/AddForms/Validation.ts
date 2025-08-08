export const ADD_ACCOUNT_VALIDATION = {
    accountNumber: {
        required: "*Account number is required",
        minLength: {
            value: 2,
            message: "*Account number must have at least 2 digits",
        },
        maxLength: {
            value: 12,
            message: "*Account number only has at most 12 digits",
        }
    },
    name: { required: "*Account name is required" },
    institution: { required: "*Institution is required" },
    balance: {
        required: "*Balance is required",
        min: {
            value: 0,
            message: "*Balance must be greater than 0",
        }
    },
    creditLimit: {
        required: "*Credit limit is required",
        min: {
            value: 0,
            message: "*Credit limit must be greater than 0",
        }
    },
    dueDate: {
        required: "*Due data is required"
    }
};

export const ADD_BILL_VALIDATION = {
    description: {
        required: "*Description is required",
    },
    amount: {
        required: "*Amount is required",
        min: {
            value: 0,
            message: "*Balance must be greater than 0",
        },
    },
};

export const ADD_STOCK_VALIDATTION = {
    corporation: {
        required: "*Corporation is required",
    },
    name: {
        required: "*Name is required",
    },
    symbol: {
        required: "*Symbol is required",
    },
    shares: {
        required: "*Shares are required",
        min: {
            value: 1,
            message: "*Shares must be greater than 1",
        },
        max: {
            value: 1000000,
            message: "*Shares must be less than 1000000",
        }
    }
}