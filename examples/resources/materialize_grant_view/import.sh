#Grants can be imported using the concatenation of GRANT, the object type, the id of the object, the id of the role and the privilege 
terraform import materialize_grant_view.example GRANT|VIEW|<view_id>|<role_id>|<privilege>
