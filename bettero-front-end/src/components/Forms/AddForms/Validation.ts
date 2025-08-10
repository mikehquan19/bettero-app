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

export const ADD_TRANSACTION_VALIDATION = {
    // validation on the description field 
    description: {
        required: "Description is required",
        maxLength: {
            value: 200,
            message: "Description must be less than 200 characters",
        }
    },
    // validation on the amount field
    amount: {
        required: "Amount is required",
        min: {
            value: 0.01,
            message: "Amount must be at least 0.01",
        },
        maxLength: {
            value: 10,
            message: "Amount must be less than 10 digits "
        }
    }
}

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

/**
 * Get the validation for the form to add or update budget plans in runtime
 */
export function getAddBudgetPlanValidation() {
    const categories = ["Total", "Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others"];
    const registerOptions = {
        recurringIncome: {
            required: "*Income is required",
            min: {
                value: 0.01,
                message: "*Income must be greater than 0",
            },
            maxLength: {
                value: 12,
                message: "*Income only has at most 12 digits",
            }
        },
    } as Record<string, any>;

    categories.forEach(category => {
        const field = category === 'Total' ? 'portionForExpense' : category;
        registerOptions[field] = {
            required: `*${category} required`,
            min: {
                value: 0,
                message: "*Percentage must be greater than or equal 0",
            },
            max: {
                value: 100,
                message: "*Percentage must be less than or equal 100",
            }
        }
    });
    return registerOptions;
}