bouncerApp.controller('ConfigurationController', function($scope, $http, $window, $timeout, $route, benchmarking, configuration) {

    $scope.backendServerCount = 1;
    $scope.selection = {};
    $scope.currentConfig = {};
    $scope.addAlertMessage = false;
    $scope.updateAlertMessage = false;
    $scope.updateMode = false;

    configuration.getAllConfigs($scope);
    $scope.addConfiguration = function(config){
        var backendServerCount = $window.document.getElementsByClassName("backendServerInput").length;
        console.log(config);
        config.concurrency = config.concurrency.toString();
        config.reqPerSecond = config.reqPerSecond.toString();
        configuration.addConfiguration($scope, config, backendServerCount);
        configuration.getAllConfigs($scope);
    };

    $scope.removeConfiguration = function(config){
        configuration.removeConfiguration($scope, config);
        configuration.getAllConfigs($scope);
    };

    $scope.updateConfiguration = function(config){
        var backendServerCount = $window.document.getElementsByClassName("backendServerInput").length;
        this.config.concurrency = parseInt(this.config.concurrency);
        this.config.reqPerSecond = parseInt(this.config.reqPerSecond);
        console.log(this.config);
        configuration.updateConfiguration($scope, this.config, backendServerCount);
    };

    $scope.selectRemoveTab = function(){
        configuration.getAllConfigs($scope);
    };

    $scope.reload = function(){
        $route.reload();
    }

    $scope.configurationSelected = function(config){
        console.log(this.config);
        console.log(this.selection);

        $scope.updateMode = true;
        if(this.selection != null){
            $scope.currentConfig = this.selection;
            this.config.concurrency = this.selection.MaxConcurrentPerBackendServer;
            this.config.targetPath = this.selection.TargetPath;
            this.config.path = this.selection.Path;
            this.config.reqPerSecond = this.selection.ReqPerSecond;
            var backServerInputCount = document.getElementById("backend-servers").childElementCount - 1;


            if(backServerInputCount<this.selection.BackendServers.length){
                for(var i=0; i<this.selection.BackendServers.length - backServerInputCount; i++){
                    $timeout(function() {
                        angular.element('#add-input-button').triggerHandler('click');
                    }, 1);    
                }        
            } else if(backServerInputCount>this.selection.BackendServers.length){
                for(var i=0; i<backServerInputCount - this.selection.BackendServers.length; i++){
                    $timeout(function() {
                        angular.element('#remove-input-button').triggerHandler('click');
                    }, 1);    
                }      
            }

            for(var i=1; i<=this.selection.BackendServers.length; i++){
                var server = "backendServer" + i;
                this.config[server] = this.selection.BackendServers[i-1].Host
            }
        }
    };

});