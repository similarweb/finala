class NumberUtils { 
    
    Format(number, fix){
        let num = number
        if (fix){
            num = num.toFixed(fix)
        }
        return num.toString().replace(/(\d)(?=(\d{3})+(?!\d))/g, '$1,')
    }
}

export default new NumberUtils();