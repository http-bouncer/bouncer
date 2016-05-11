bouncerApp.factory('configuration', function ($http, $interval, $timeout) {


  var getBackendServers = function($scope,config, backendServerCount){
      var backendServers = [];

      for(var i=1; i<=backendServerCount; i++){
            var serverHostName = "backendServer" + i;
            var backendServer = {
              "host": config[serverHostName]
            };
            backendServers.push(backendServer);
            console.log(backendServer);
        }
          return backendServers;
    };

   return {

       addConfiguration: function($scope, config, backendServerCount){
        var url = "http://localhost:8080/addConfiguration";
        if(config.targetPath == undefined) config.targetPath = ""
        if(config.path == undefined) config.path = ""
        var config = {
            "host": config.hostName,
            "path": config.path,
            "targetPath": config.targetPath,
            "reqPerSecond": config.reqPerSecond,
            "concurrency": config.concurrency,
            "backendServers": getBackendServers($scope, config, backendServerCount)
        };

        console.log(config);

        $http({
            method: 'POST',
            url: url,
            headers: {'Content-Type': 'application/json'}, 
            data: JSON.stringify(config)
        }).success(function (data) {
            console.log(JSON.stringify(data));
            $scope.addAlertMessage = true;
            $timeout(function () { $scope.addAlertMessage = false; }, 3000);   
        });
       },

       updateConfiguration: function($scope, config, backendServerCount){
        console.log(config);
        console.log(backendServerCount);
          var url = "http://localhost:8080/updateConfiguration";
          if(config.targetPath == undefined) config.targetPath = ""
          if(config.path == undefined) config.path = ""

          var config = {
              "host": config.hostName,
              "path": config.path,
              "targetPath": config.targetPath,
              "reqPerSecond": config.reqPerSecond,
              "concurrency": config.concurrency,
              "backendServers": getBackendServers($scope, config, backendServerCount)
          };

          console.log(config.backendServers)

          $http({
              method: 'POST',
              url: url,
              headers: {'Content-Type': 'application/json'}, 
              data: JSON.stringify(config)
          }).success(function (data) {
              console.log(JSON.stringify(data));
              $scope.updateAlertMessage = true;
              $timeout(function () { $scope.updateAlertMessage = false; }, 3000);   
          });
       },

       removeConfiguration: function($scope, config){
        console.log(config);
          var url = "http://localhost:8080/removeConfiguration";
          $http({
              method: 'POST',
              url: url,
              headers: {'Content-Type': 'application/json'}, 
              data: JSON.stringify(config)
          }).success(function (data) {
              console.log(JSON.stringify(data));
          });
       },

       getAllConfigs: function($scope){
        var url = "http://localhost:8080/getAllConfigs";
        $http({
            method: 'POST',
            url: url,
            headers: {'Content-Type': 'application/json'}
        }).success(function (data) {
            console.log(JSON.stringify(data));
            console.log(data);
            if(data.length < 1 || data == "null" || data == null) $scope.configs = [];
            else $scope.configs = data;
        });
       }
   }
});