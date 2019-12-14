class TextUtils { 
    
    ParseName(sourceName){
        let str = sourceName.replace(/_/g, " ").replace(/aws/g, "AWS")
        return str
    }

    CapitalizeWords(str){
        return str.replace(/\w\S*/g, function(txt){return txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase();});
    }
}

export default new TextUtils();