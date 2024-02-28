package provider

// func TestAccUser_basic(t *testing.T) {
// 	resourceName := "materialize_user.test"
// 	email := "test@example.com"
// 	role := "Member" // Assuming this is a valid role in your system

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:          func() { testAccPreCheck(t) },
// 		ProviderFactories: testAccProviderFactories,
// 		CheckDestroy:      testAccCheckUserDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccUserConfigBasic(email, role),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr(resourceName, "email", email),
// 					resource.TestCheckResourceAttr(resourceName, "roles.0", role),
// 					resource.TestCheckResourceAttrSet(resourceName, "auth_provider"),
// 					resource.TestCheckResourceAttrSet(resourceName, "verified"),
// 					resource.TestCheckResourceAttrSet(resourceName, "metadata"),
// 				),
// 			},
// 		},
// 	})
// }

// func testAccUserConfigBasic(email, role string) string {
// 	return fmt.Sprintf(`
//         resource "materialize_user" "test" {
//             email = "%s"
//             roles = ["%s"]
//         }
//     `, email, role)
// }

// func testAccCheckUserExists(resourceName string) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		meta := testAccProvider.Meta()
// 		providerMeta, _ := utils.GetProviderMeta(meta)
// 		client := providerMeta.Frontegg
// 		rs, ok := s.RootModule().Resources[resourceName]
// 		if !ok {
// 			return fmt.Errorf("Not found: %s", resourceName)
// 		}

// 		userID := rs.Primary.ID
// 		_, err := frontegg.ReadUser(context.Background(), client, userID)
// 		if err != nil {
// 			return fmt.Errorf("Error fetching user with ID [%s]: %s", userID, err)
// 		}

// 		return nil
// 	}
// }

// func testAccCheckUserDestroy(s *terraform.State) error {
// 	for _, rs := range s.RootModule().Resources {
// 		if rs.Type != "materialize_user" {
// 			continue
// 		}

// 		meta := testAccProvider.Meta()
// 		providerMeta, _ := utils.GetProviderMeta(meta)
// 		client := providerMeta.Frontegg

// 		_, err := frontegg.ReadUser(context.Background(), client, rs.Primary.ID)
// 		if err == nil {
// 			return fmt.Errorf("User %s still exists", rs.Primary.ID)
// 		}
// 	}

// 	return nil
// }
