package main

import (
	"fmt"
	"github.com/vulkan-go/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
)

type appObject struct {
	window   *glfw.Window
	instance vk.Instance
	// Surface Specific
	surface vk.Surface
	//surfaceFormats []vk.SurfaceFormat
	surfaceFormat    vk.SurfaceFormat
	displaySize      vk.Extent2D
	displayFormat    vk.Format
	swapchains       []vk.Swapchain
	swapchainslength []uint32
	imageViews       []vk.ImageView
	//Device Specific
	logicalDevice    vk.Device
	physicalDevices  []vk.PhysicalDevice
	graphicsQueuePtr *vk.Queue
	graphicsQueueIdx []uint32
	//Command Buffer Specific
	commandPool    vk.CommandPool
	commandBuffers []vk.CommandBuffer
}

// type deviceInfo struct {
// 	gpus     []vk.PhysicalDevice
// 	instance vk.Instance
// 	surface  vk.Surface
// 	device   vk.Device
// }

// Vulkan Device Queue family capabilities Enum
// https://www.khronos.org/registry/vulkan/specs/1.2-extensions/man/html/VkQueueFlagBits.html
type VkQueueFlagBits uint32

const (
	VK_QUEUE_GRAPHICS_BIT VkQueueFlagBits = 1 << iota
	VK_QUEUE_COMPUTE_BIT
	VK_QUEUE_TRANSFER_BIT
	VK_QUEUE_SPARSE_BINDING_BIT
	VK_QUEUE_PROTECTED_BIT
	VK_QUEUE_FLAG_BITS_MAX_ENUM = 0xFFFFFFFF >> 1
)

// Device Queue Capabilities
// https://www.khronos.org/registry/vulkan/specs/1.2-extensions/man/html/VkQueueFlagBits.html

func main() {
	//-------------------------
	// Initialization Stage
	//-------------------------

	var app appObject

	orPanic(glfw.Init())
	// GetVulkanGetInstanceProcAddress returns the function pointer used to find Vulkan core or
	// extension functions. The return value of this function can be passed to the Vulkan library.
	// Note that this function does not work the same way as the glfwGetInstanceProcAddress.
	// https://github.com/vulkan-go/glfw/blob/v3.3/v3.3/glfw/vulkan.go
	// SetGetInstanceProcAddr sets the GetInstanceProcAddr function pointer used to load Vulkan symbols.
	// This function must be called before vulkan.Init()
	// https://godoc.org/github.com/vulkan-go/vulkan#SetGetInstanceProcAddr
	//fmt.Println(glfw.GetVulkanGetInstanceProcAddress())
	vk.SetGetInstanceProcAddr(glfw.GetVulkanGetInstanceProcAddress())
	orPanic(vk.Init())

	// *** 1 List Layers available ***//

	//xPrintInstanceLayerProperties()
	//xPrintInstanceExtensionProperties()

	// *** 2 GLFW Window Creation ***//
	// Vulkan WSI extensions avaibale for different platforms
	// https://www.glfw.org/docs/latest/compat.html
	app.window = xCreateWindowGLFW()

	// *** 3 Instance and Application Creation ***//

	app.instance = xCreateInstance()
	fmt.Println(app.instance)

	//xDevicesInfo(instance)
	app.physicalDevices, _ = xGetDevices(app.instance)
	xGetDeviceQueueFamilyProperties(app.physicalDevices[0])

	app.logicalDevice, _ = xCreateLogicalDevice(app.instance)

	// Search for Graphics queue that is capable for supporting Graphics Operations
	xCreateSurface(&app)
	//graphicsQueueIndex := xGetGPUQueueSupportingGraphicsOps()
	app.graphicsQueueIdx = xGetGPUQueueSupportingGraphicsOps(&app)
	xGetSurfaceFormats(&app)
	xCreateSwapChain(&app)
	xCommandBufferInitialization(&app)
	xCreateImageView(&app)

	// Cleanup task
	for _, imageView := range app.imageViews {
		vk.DestroyImageView(app.logicalDevice, imageView, nil)
	}
	vk.DestroyCommandPool(app.logicalDevice, app.commandPool, nil)
	vk.DestroySwapchain(app.logicalDevice, app.swapchains[0], nil)
	vk.DestroySurface(app.instance, app.surface, nil)
	app.window.Destroy()
	vk.DestroyDevice(app.logicalDevice, nil)
	vk.DestroyInstance(app.instance, nil)
}

// func xAllocateCmdBufferFromCmdPool(app *appObject) {
// 	AllocateCommandBuffers(app.physicalDevices[0], pAllocateInfo *CommandBufferAllocateInfo, pCommandBuffers []CommandBuffer)
// }

//  Create the image view of the retrieved swapchain images
func xCreateImageView(app *appObject) {
	var swapchainImageCount uint32 // If this is populated with '2' by below function, then it means swap chain supports double buffering
	err := vk.Error(vk.GetSwapchainImages(app.logicalDevice, app.swapchains[0], &swapchainImageCount, nil))
	if err != nil {
		err = fmt.Errorf("vkCreateDevice failed with %s", err)
		return
	}
	var swapchainImages = make([]vk.Image, swapchainImageCount)
	err = vk.Error(vk.GetSwapchainImages(app.logicalDevice, app.swapchains[0], &swapchainImageCount, swapchainImages))
	if err != nil {
		err = fmt.Errorf("vkCreateDevice failed with %s", err)
		return
	}
	//app.swapchainImages = swapchainImages
	//fmt.Println(swapchainImageCount)
	for _, image := range swapchainImages {
		var imageView vk.ImageView
		imageViewCreateInfo := vk.ImageViewCreateInfo{
			SType:    vk.StructureTypeImageViewCreateInfo,
			Image:    image,
			ViewType: vk.ImageViewType2d,
			Format:   app.surfaceFormat.Format,
			Components: vk.ComponentMapping{
				R: vk.ComponentSwizzleR,
				G: vk.ComponentSwizzleG,
				B: vk.ComponentSwizzleB,
				A: vk.ComponentSwizzleA,
			},
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
				LevelCount: 1,
				LayerCount: 1,
			},
		}
		err := vk.Error(vk.CreateImageView(app.logicalDevice, &imageViewCreateInfo, nil, &imageView))
		if err != nil {
			err = fmt.Errorf("ImageView creation failed with %s", err)
			return
		}
		app.imageViews = append(app.imageViews, imageView)
	}
	fmt.Println("Created Image View......")
}

func xCommandBufferInitialization(app *appObject) {
	var commandPool vk.CommandPool
	cmdPoolCreateInfo := vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
		QueueFamilyIndex: 0,
	}
	err := vk.Error(vk.CreateCommandPool(app.logicalDevice, &cmdPoolCreateInfo, nil, &commandPool))
	if err != nil {
		err = fmt.Errorf("vkCreateDevice failed with %s", err)
		return
	}
	app.commandPool = commandPool
	fmt.Println("Created Command Pool..........")
	var commandBuffers = make([]vk.CommandBuffer, app.swapchainslength[0])
	var cmdBufferAllocateInfo = vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        commandPool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: app.swapchainslength[0],
	}
	err = vk.Error(vk.AllocateCommandBuffers(app.logicalDevice, &cmdBufferAllocateInfo, commandBuffers))
	if err != nil {
		err = fmt.Errorf("vk.AllocateCommandBuffers failed with %s", err)
		return
	}
	app.commandBuffers = commandBuffers
	fmt.Println("Created Command Buffer..........")
}

func xCreateSwapChain(app *appObject) {
	var surfaceCapabilities vk.SurfaceCapabilities
	err := vk.Error(vk.GetPhysicalDeviceSurfaceCapabilities(app.physicalDevices[0], app.surface, &surfaceCapabilities))
	if err != nil {
		err = fmt.Errorf("Failed getting surface capabilities with error %s", err)
		return
	}
	surfaceCapabilities.Deref()
	app.displaySize = surfaceCapabilities.CurrentExtent
	app.displaySize.Deref()
	app.displayFormat = app.surfaceFormat.Format
	swapChainCreateInfo := vk.SwapchainCreateInfo{
		SType:            vk.StructureTypeSwapchainCreateInfo,
		Surface:          app.surface,
		MinImageCount:    surfaceCapabilities.MinImageCount,
		ImageFormat:      app.surfaceFormat.Format,
		ImageColorSpace:  app.surfaceFormat.ColorSpace,
		ImageExtent:      surfaceCapabilities.CurrentExtent,
		ImageUsage:       vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit),
		PreTransform:     vk.SurfaceTransformIdentityBit,
		ImageArrayLayers: 1, //Teels about whether it's virtual 3D view or not - imageArrayLayers is the number of views in a multiview/stereo surface. For non-stereoscopic-3D applications, this value is 1.
		// https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/html/vkspec.html#VkSwapchainCreateInfoKHR
		ImageSharingMode:      vk.SharingModeExclusive,
		QueueFamilyIndexCount: 1,
		PQueueFamilyIndices:   app.graphicsQueueIdx,
		PresentMode:           vk.PresentModeFifo,
		OldSwapchain:          vk.NullSwapchain,
		Clipped:               vk.False,
		CompositeAlpha:        vk.CompositeAlphaOpaqueBit,
	}
	var swapChains = make([]vk.Swapchain, 1)
	var swapchainlength = make([]uint32, 1)
	err = vk.Error(vk.CreateSwapchain(app.logicalDevice, &swapChainCreateInfo, nil, &swapChains[0]))
	if err != nil {
		err = fmt.Errorf("vk.CreateSwapchain failed with %s", err)
		return
	}
	app.swapchains = swapChains
	app.swapchainslength = swapchainlength
	err = vk.Error(vk.GetSwapchainImages(app.logicalDevice, swapChains[0], &swapchainlength[0], nil))
	if err != nil {
		err = fmt.Errorf("vk.GetSwapchainImages failed with %s", err)
		return
	}
	fmt.Println("Create Swapchain.......")
}

func xGetSurfaceFormats(app *appObject) {
	var graphicQueue vk.Queue
	var formatCount uint32
	vk.GetDeviceQueue(app.logicalDevice, app.graphicsQueueIdx[0], 0, &graphicQueue)
	app.graphicsQueuePtr = &graphicQueue
	vk.GetPhysicalDeviceSurfaceFormats(app.physicalDevices[0], app.surface, &formatCount, nil)
	var surfaceformats = make([]vk.SurfaceFormat, formatCount)
	vk.GetPhysicalDeviceSurfaceFormats(app.physicalDevices[0], app.surface, &formatCount, surfaceformats)
	surfaceformats[0].Deref()
	for i := 0; i < int(formatCount); i++ {
		if surfaceformats[i].Format == vk.FormatB8g8r8a8Unorm || surfaceformats[i].Format == vk.FormatR8g8b8a8Unorm {
			app.surfaceFormat = surfaceformats[i]
			break
		}
	}
	//app.surfaceFormats = surfaceformats
	fmt.Println("Retrieved SurfaceFormats.......")
}

func xGetGPUQueueSupportingGraphicsOps(app *appObject) []uint32 {
	var familyPropertyCount uint32
	var isPresentationSuported vk.Bool32
	var index = make(map[uint32]uint32)
	var uniquekeys []uint32
	// Get list of Queue family count
	vk.GetPhysicalDeviceQueueFamilyProperties(app.physicalDevices[0], &familyPropertyCount, nil)
	for i := 0; i < int(familyPropertyCount); i++ {
		var idx uint32
		vk.GetPhysicalDeviceSurfaceSupport(app.physicalDevices[0], idx, app.surface, &isPresentationSuported)
		if isPresentationSuported == 1 {
			index[idx] = index[idx] + 1
		}
	}
	for k := range index {
		uniquekeys = append(uniquekeys, k)
	}
	return uniquekeys
}

func xCreateSurface(app *appObject) {
	surfacePtr, err := app.window.CreateWindowSurface(app.instance, nil)
	if err != nil {
		fmt.Println("Error creating windows surface ", err)
		app.surface = vk.NullSurface
	}
	app.surface = vk.SurfaceFromPointer(surfacePtr)
}

func xCreateWindowGLFW() *glfw.Window {
	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	window, err := glfw.CreateWindow(800, 600, "My Game Engine", nil, nil)
	if err != nil {
		fmt.Println("Failed to create window with error ", err)
		return nil
	}
	fmt.Println("Created GLFW Window.......")
	return window
}

func xCreateLogicalDevice(instance vk.Instance) (vk.Device, error) {
	gpudevices, err := xGetDevices(instance)
	if err != nil {
		err = fmt.Errorf("Failed to get list of physical devices with error: %s", err)
		return nil, err
	}
	// https://www.khronos.org/registry/vulkan/specs/1.2-extensions/html/vkspec.html#VkDeviceQueueCreateInfo
	// See output of 'xGetDeviceQueueFamilyProperties()' to see more details
	// var deviceQueueCreateInfoSlice []vk.DeviceQueueCreateInfo = []vk.DeviceQueueCreateInfo{
	// 	{
	// 		SType:            vk.StructureTypeDeviceQueueCreateInfo,
	// 		QueueFamilyIndex: 0,
	// 		QueueCount:       1,
	// 		PQueuePriorities: []float32{1.0},
	// 		Flags:            0x7FFFFFFF, // This is equivalent of 'VK_DEVICE_QUEUE_CREATE_FLAG_BITS_MAX_ENUM' bit set or in other words, enable everything available
	// 	},
	// 	{
	// 		SType:            vk.StructureTypeDeviceQueueCreateInfo,
	// 		QueueFamilyIndex: 1,
	// 		QueueCount:       1,
	// 		PQueuePriorities: []float32{1.0},
	// 		Flags:            0x7FFFFFFF,
	// 	},
	// }
	deviceQueueCreateInfoSlice := []vk.DeviceQueueCreateInfo{{
		SType:            vk.StructureTypeDeviceQueueCreateInfo,
		QueueCount:       1,
		PQueuePriorities: []float32{1.0},
	}}
	//var deviceExtensions = []string{"VK_KHR_surface\x00"}
	var deviceExtensions = []string{"VK_KHR_swapchain\x00"}
	var deviceLayers = []string{"VK_LAYER_KHRONOS_validation\x00"}
	var deviceCreateInfo *vk.DeviceCreateInfo = &vk.DeviceCreateInfo{
		SType:                   vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount:    uint32(len(deviceQueueCreateInfoSlice)),
		PQueueCreateInfos:       deviceQueueCreateInfoSlice,
		EnabledLayerCount:       uint32(len(deviceLayers)),
		PpEnabledLayerNames:     deviceLayers,
		EnabledExtensionCount:   uint32(len(deviceExtensions)),
		PpEnabledExtensionNames: deviceExtensions,
	}
	var logicalDevice vk.Device
	err = vk.Error(vk.CreateDevice(gpudevices[0], deviceCreateInfo, nil, &logicalDevice))
	if err != nil {
		err = fmt.Errorf("vkCreateDevice failed with %s", err)
		return nil, err
	}
	fmt.Println("Created Logical Device.......")
	return logicalDevice, nil
}

func xGetDeviceQueueFamilyProperties(device vk.PhysicalDevice) {
	var familyPropertyCount uint32
	//var gpuqueuecapabilities VkQueueFlagBits
	var gpuqueuecapabilities uint32
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &familyPropertyCount, nil)
	fmt.Println(familyPropertyCount, " Queue Family/Families found on GPU")
	var gpuQueueFamilyProperties = make([]vk.QueueFamilyProperties, familyPropertyCount)
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &familyPropertyCount, gpuQueueFamilyProperties)
	fmt.Println("Following GPU Queue Capabilities were found ....")
	for idx, q := range gpuQueueFamilyProperties {
		q.Deref()
		//f := vk.ToString(q.QueueFlags)
		//fmt.Printf("%v\t: %v\n", idx+1, f)
		fmt.Printf("Flag(s) value %v is(are) set for Queue %v\n", q.QueueFlags, idx)
		gpuqueuecapabilities = uint32(q.QueueFlags)
		if gpuqueuecapabilities&uint32(VK_QUEUE_GRAPHICS_BIT) != 0 {
			fmt.Println("\tVK_QUEUE_GRAPHICS_BIT ")
		}
		if gpuqueuecapabilities&uint32(VK_QUEUE_COMPUTE_BIT) != 0 {
			fmt.Println("\tVK_QUEUE_COMPUTE_BIT ")
		}
		if gpuqueuecapabilities&uint32(VK_QUEUE_TRANSFER_BIT) != 0 {
			fmt.Println("\tVK_QUEUE_TRANSFER_BIT ")
		}
		if gpuqueuecapabilities&uint32(VK_QUEUE_SPARSE_BINDING_BIT) != 0 {
			fmt.Println("\tVK_QUEUE_SPARSE_BINDING_BIT ")
		}
		if gpuqueuecapabilities&uint32(VK_QUEUE_PROTECTED_BIT) != 0 {
			fmt.Println("\tVK_QUEUE_PROTECTED_BIT ")
		}
		if gpuqueuecapabilities&uint32(VK_QUEUE_FLAG_BITS_MAX_ENUM) != 0 {
			fmt.Println("\tVK_QUEUE_FLAG_BITS_MAX_ENUM ")
		}
	}
	fmt.Println("Retrieved GPU Graphics Queue information.......")
}

func xGetDevices(instance vk.Instance) ([]vk.PhysicalDevice, error) {

	var deviceCount uint32
	err := vk.Error(vk.EnumeratePhysicalDevices(instance, &deviceCount, nil))
	if err != nil {
		err = fmt.Errorf("Failed to get list of physical devices with error: %s", err)
		return nil, err
	}
	fmt.Println(deviceCount, " Physical Device(s) found")
	var devices = make([]vk.PhysicalDevice, deviceCount)
	err = vk.Error(vk.EnumeratePhysicalDevices(instance, &deviceCount, devices))
	if err != nil {
		err = fmt.Errorf("Failed to get properties of physical devices with error: %s", err)
		return nil, err
	}
	return devices, nil
}

func xCreateInstance() vk.Instance {
	var appInfo *vk.ApplicationInfo = &vk.ApplicationInfo{
		SType:              vk.StructureTypeApplicationInfo,
		PNext:              nil,
		PApplicationName:   "myVulkan Application\x00",
		ApiVersion:         vk.MakeVersion(1, 0, 0), // Throws 'vulkan error: incompatible driver' error with incorrect version number
		ApplicationVersion: vk.MakeVersion(1, 0, 0),
		PEngineName:        "My Game Engine\x00",
		EngineVersion:      vk.MakeVersion(0, 1, 0),
	}
	var instance vk.Instance
	var layers = []string{"VK_LAYER_KHRONOS_validation\x00"}
	var extensions = []string{"VK_KHR_surface\x00", "VK_KHR_win32_surface\x00"}
	var instanceInfo = vk.InstanceCreateInfo{
		SType: vk.StructureTypeInstanceCreateInfo,
		PNext: nil,
		//Flags:                   nil,
		PApplicationInfo:        appInfo,
		EnabledLayerCount:       uint32(len(layers)),
		PpEnabledLayerNames:     layers,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
	}
	orPanic(vk.CreateInstance(&instanceInfo, nil, &instance))
	//InitInstance obtains instance PFNs for Vulkan API functions, this is necessary on OS X using MoltenVK, but for the other platforms it's an option. Not implemented for Android.
	vk.InitInstance(instance)
	return instance
}

func xPrintInstanceExtensionProperties() {
	var extensionCount uint32
	vk.EnumerateInstanceExtensionProperties("", &extensionCount, nil)
	fmt.Println(extensionCount, " Extensions found\n")
	extensions := make([]vk.ExtensionProperties, extensionCount)
	vk.EnumerateInstanceExtensionProperties("", &extensionCount, extensions)
	for idx, e := range extensions {
		e.Deref()
		var name string = vk.ToString(e.ExtensionName[:])
		fmt.Printf("%v\t: %v\n", idx+1, name)
	}
}

func xPrintInstanceLayerProperties() {
	var layerCount uint32
	vk.EnumerateInstanceLayerProperties(&layerCount, nil)
	fmt.Println(layerCount, " Layers found\n")
	instanceLayers := make([]vk.LayerProperties, layerCount)
	vk.EnumerateInstanceLayerProperties(&layerCount, instanceLayers)
	for idx, lp := range instanceLayers {
		lp.Deref()
		var name string = vk.ToString(lp.LayerName[:])
		var desc string = vk.ToString(lp.Description[:])
		fmt.Println("Layer ", idx, ": ", name)
		fmt.Println(" Description ", desc)
	}
}

func orPanic(err interface{}) {
	switch v := err.(type) {
	case error:
		if v != nil {
			panic(err)
		}
	case vk.Result:
		if err := vk.Error(v); err != nil {
			panic(err)
		}
	case bool:
		if !v {
			panic("condition failed: != true")
		}
	}
}
